package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"golang.org/x/net/context"

	coap "github.com/coapcloud/go-coap"
	_ "github.com/joho/godotenv/autoload"

	firebase "firebase.google.com/go"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var docs map[int]*firestore.DocumentSnapshot

func init() {
	docs = make(map[int]*firestore.DocumentSnapshot)
}

func run(r *Router) {
	b, err := base64.StdEncoding.DecodeString(os.Getenv("FIREBASE_SERVICE_ACCOUNT"))
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	opt := option.WithCredentialsJSON(b)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	client, err := app.Firestore(context.Background())
	if err != nil {
		panic(err)
	}

	q := client.Collection("endpoints")
	iter := q.Snapshots(context.Background())

	for {
		snap, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			log.Println(err)
		}

		for _, v := range snap.Changes {
			switch v.Kind {
			case firestore.DocumentAdded:
				fmt.Println("trying to add route")
				hotAddRoute(r, v.Doc)
				docs[v.NewIndex] = v.Doc
			case firestore.DocumentModified:
				if err := hotModRoute(r, v.Doc); err != nil {
					log.Println(err)
					continue
				}
				docs[v.NewIndex] = v.Doc
			case firestore.DocumentRemoved:
				if err := hotDelRoute(r, v.Doc); err != nil {
					log.Println(err)
					continue
				}
				delete(docs, v.NewIndex)
			}
		}

		printdocs()
	}
}

func printdocs() {
	for _, v := range docs {
		fmt.Printf("%+v\n", v.Data())
	}
}

func hotAddRoute(r *Router, d *firestore.DocumentSnapshot) {
	path, funcName := getVarsFromData(d)

	r.HotRegisterRoute(coap.POST, path, funcName)
}
func hotModRoute(r *Router, d *firestore.DocumentSnapshot) error {
	path, funcName := getVarsFromData(d)

	return r.HotModifyRoute(coap.POST, path, funcName)
}
func hotDelRoute(r *Router, d *firestore.DocumentSnapshot) error {
	path, funcName := getVarsFromData(d)

	return r.HotDeRegisterRoute(coap.POST, path, funcName)
}

func getVarsFromData(d *firestore.DocumentSnapshot) (string, string) {
	data := d.Data()

	path, ok := data["path"]
	if !ok {
		log.Println("path for route not found")
		return "", ""
	}

	pathStr, ok := path.(string)
	if !ok {
		log.Println("type assertion for route failed")
		return "", ""
	}

	funcName, ok := data["function"]
	if !ok {
		log.Println("type assertion for route failed")
		return "", ""
	}

	funcNameStr, ok := funcName.(string)
	if !ok {
		log.Println("type assertion for route failed")
		return "", ""
	}

	return pathStr, funcNameStr
}
