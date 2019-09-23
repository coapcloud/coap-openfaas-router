package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/derekparker/trie"
	coap "github.com/go-ocf/go-coap"
)

// Router Root-Level Router
type Router struct {
	*trie.Trie
}

// NewRouter ...
func NewRouter() Router {
	return Router{
		trie.New(),
	}
}

// ServeCOAP - implementation of coap.Handler
func (r Router) ServeCOAP(w coap.ResponseWriter, req *coap.Request) {
	var (
		respBdy string
		err     error
	)
	log.Printf("Got message: %#v path=%q: from %v\n", req.Msg.PathString(), req.Msg, req.Client.RemoteAddr())

	funcID, ok := r.match(req.Msg.Code(), req.Msg.PathString())
	if !ok {
		log.Println("could not match route")
		w.SetCode(coap.NotFound)
		respBdy = fmt.Sprintf("not found")
	}

	// run openfaas function for route + verb
	respBdy, err = openfaasCall(funcID, req.Msg.Payload())
	if err != nil {
		log.Printf("Error while trying to invoke openfaas function %v\n", err)
		w.SetCode(coap.InternalServerError)
		respBdy = fmt.Sprint("could not run callback for request")
	}

	ctx, cancel := context.WithTimeout(req.Ctx, 3*time.Second)
	defer cancel()

	log.Printf("Writing response to %v\n\n", req.Client.RemoteAddr())
	w.SetContentFormat(coap.TextPlain)
	if _, err := w.WriteWithContext(ctx, []byte(respBdy)); err != nil {
		log.Printf("Cannot send response: %v", err)
	}
}

func openfaasCall(funcID string, bdy []byte) (string, error) {
	resp, err := http.Post(fmt.Sprintf("%s/function/%s", openfaasAddress, funcID), "application/octet-stream", bytes.NewBuffer(bdy))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	output := string(b)
	return fmt.Sprintf("Invoked func: %s with input: %v. Result: %s", funcID, string(bdy), output), nil
}

func (r *Router) registerRoute(verb coap.COAPCode, path, openfaasFuncID string) {
	key := routeKey(verb, path)

	log.Printf("registering route: %v %v -> openfaas func: %s\n", verb.String(), path, openfaasFuncID)

	node := r.Add(key, openfaasFuncID)
	if node != nil {
		log.Printf("registered route: %s /%s to func %q\n", verb.String(), path, openfaasFuncID)
	}
}

func (r *Router) match(verb coap.COAPCode, path string) (string, bool) {
	key := routeKey(verb, path)

	fmt.Println(key)

	node, ok := r.Find(key)
	if !ok {
		log.Printf("couldn't find openfaas function id for route: %s\n", key)
		return "", false
	}

	meta := node.Meta()

	v, ok := meta.(string)
	if !ok {
		log.Printf("couldn't find string-ey openfaas function id for route: %s\n", key)
		return "", false
	}

	return v, true
}

func routeKey(verb coap.COAPCode, path string) string {
	return fmt.Sprintf("%d-%s", verb, path)
}

// GET - register a CoAP GET /{path} to a func callback
func (r *Router) GET(path, openfaasFuncID string) {
	if r != nil {
		r.registerRoute(coap.GET, path, openfaasFuncID)
	}

	log.Println("can't register route to nil router")
}

// POST - register a CoAP POST /{path} to a func callback
func (r *Router) POST(path, openfaasFuncID string) {
	if r != nil {
		r.registerRoute(coap.POST, path, openfaasFuncID)
	}

	log.Println("can't register route to nil router")
}

// PUT - register a CoAP PUT /{path} to a func callback
func (r *Router) PUT(path, openfaasFuncID string) {
	if r != nil {
		r.registerRoute(coap.PUT, path, openfaasFuncID)
	}

	log.Println("can't register route to nil router")
}

// DELETE - register a CoAP DELETE /{path} to a func callback
func (r *Router) DELETE(path, openfaasFuncID string) {
	if r != nil {
		r.registerRoute(coap.DELETE, path, openfaasFuncID)
	}

	log.Println("can't register route to nil router")
}
