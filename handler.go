package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	coap "github.com/go-ocf/go-coap"
)

type openfaasCallback string

type faasHandler struct {
	OpenFaasFuncID openfaasCallback
	Code           coap.COAPCode
	Type           coap.COAPType
}

func (h faasHandler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	var (
		respBdy string
		err     error
	)

	fmt.Println("sent", r.Msg.Code())
	fmt.Println("expected", h.Code)

	log.Printf("Got message: %#v path=%q: from %v\n", r.Msg.Path(), r.Msg, r.Client.RemoteAddr())

	if r.Msg.Code() != h.Code {
		w.SetCode(coap.MethodNotAllowed)
		respBdy = fmt.Sprintf("CoAP method: %q not allowed", r.Msg.Code())
	} else {
		// run openfaas function for route + verb
		respBdy, err = openfaasCall(h.OpenFaasFuncID, r.Msg.Payload())
		if err != nil {
			log.Printf("Error while trying to invoke openfaas function %v\n", err)
			w.SetCode(coap.InternalServerError)
			respBdy = fmt.Sprint("could not run callback for request")
		}
	}

	ctx, cancel := context.WithTimeout(r.Ctx, 3*time.Second)
	defer cancel()

	log.Printf("Writing response to %v\n\n", r.Client.RemoteAddr())
	w.SetContentFormat(coap.TextPlain)
	if _, err := w.WriteWithContext(ctx, []byte(respBdy)); err != nil {
		log.Printf("Cannot send response: %v", err)
	}
}

func openfaasCall(name openfaasCallback, bdy []byte) (string, error) {
	resp, err := http.Post(fmt.Sprintf("%s/function/%s", openfaasAddress, string(name)), "application/octet-stream", bytes.NewBuffer(bdy))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	output := string(b)
	return fmt.Sprintf("Invoked func: %v with input: %v. Result: %s", name, string(bdy), output), nil
}
