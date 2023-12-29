package server

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Unquabain/ephemeral/data"
	"github.com/Unquabain/ephemeral/envelope"
	"github.com/apex/log"
)

//go:embed index.html
var indexHTML []byte

type webError struct {
	error
	code          int
	publicMessage string
}

func werr(err error, code int, publicMessage string) *webError {
	return &webError{
		error:         err,
		code:          code,
		publicMessage: publicMessage,
	}
}

type getBody func(target any) error
type response interface {
	Respond(http.ResponseWriter) error
}

type textResponse []byte

func (r textResponse) Respond(w http.ResponseWriter) error {
	w.Header().Add(`Content-Type`, `text/plain`)
	_, err := w.Write(r)
	return err
}

type htmlResponse []byte

func (r htmlResponse) Respond(w http.ResponseWriter) error {
	w.Header().Add(`Content-Type`, `text/html`)
	_, err := w.Write(r)
	return err
}

type textReaderResponse struct {
	io.Reader
}

func (r textReaderResponse) Respond(w http.ResponseWriter) error {
	w.Header().Add(`Content-Type`, `text/plain`)
	_, err := io.Copy(w, r)
	return err
}

type jsonResponse struct {
	data any
}

func (r jsonResponse) Respond(w http.ResponseWriter) error {
	w.Header().Add(`Content-Type`, `application/json`)
	return json.NewEncoder(w).Encode(r.data)
}

type apiHandler func(getBody) (response, *webError)

func handlerFunc(f apiHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyInto := func(target any) error {
			defer r.Body.Close()
			if r.Method != http.MethodPost {
				return fmt.Errorf(`improper HTTP verb: %s`, r.Method)
			}
			return json.NewDecoder(r.Body).Decode(target)
		}
		if resp, err := f(bodyInto); err != nil {
			serveError(err, w, r)
		} else if err := resp.Respond(w); err != nil {
			log.WithError(err).Error(`unable to serve API content.`)
		}
	})
}

func serveError(err *webError, w http.ResponseWriter, r *http.Request) {
	log.WithError(err.error).
		WithField(`response code`, err.code).
		WithField(`url`, r.URL).
		WithField(`method`, r.Method).
		Error(`an error serving endpoint`)
	w.Header().Add(`Content-Type`, `application/json`)
	w.WriteHeader(err.code)
	json.NewEncoder(w).Encode(struct {
		Error string
	}{err.publicMessage})
}

func request(bodyInto getBody) (response, *webError) {
	var requestData struct {
		Description string
	}
	var responseData struct {
		PrivateRequest envelope.Envelope
		PublicRequest  envelope.Envelope
	}
	if err := bodyInto(&requestData); err != nil {
		return nil, werr(err, 400, `unable to parse request parameters`)
	}
	privateRequest, err := data.NewRequest(requestData.Description)
	if err != nil {
		return nil, werr(err, 500, `unable to create new request`)
	}
	responseData.PrivateRequest.Name = `PRIVATE REQUEST`
	responseData.PrivateRequest.Prelude = privateRequest.Description
	if err := responseData.PrivateRequest.Stuff(privateRequest); err != nil {
		return nil, werr(err, 500, `unable to stuff private request envelope`)
	}
	responseData.PublicRequest.Name = `PUBLIC REQUEST`
	responseData.PublicRequest.Prelude = privateRequest.Description
	if err := responseData.PublicRequest.Stuff(privateRequest.Public()); err != nil {
		return nil, werr(err, 500, `unable to stuff public request envelope`)
	}

	return jsonResponse{responseData}, nil
}

func respond(bodyInto getBody) (response, *webError) {
	var (
		requestData struct {
			PublicRequest envelope.Envelope
			Data          string
		}
		publicRequest    data.PublicRequest
		responseEnvelope envelope.Envelope
	)
	if err := bodyInto(&requestData); err != nil {
		return nil, werr(err, 400, `unable to understand request parameters`)
	}
	if err := requestData.PublicRequest.Open(&publicRequest); err != nil {
		return nil, werr(err, 400, `unable to understand public request`)
	}
	responseEnvelope.Name = `RESPONSE`
	responseEnvelope.Prelude = publicRequest.Description

	if response, err := publicRequest.Encode([]byte(requestData.Data)); err != nil {
		return nil, werr(err, 500, `unable to encode response`)
	} else if err := responseEnvelope.Stuff(response); err != nil {
		return nil, werr(err, 500, `unable to stuff response envelope`)
	}
	return textReaderResponse{responseEnvelope.Reader()}, nil
}

func receive(bodyInto getBody) (response, *webError) {
	var (
		requestData struct {
			PrivateRequest envelope.Envelope
			Data           envelope.Envelope
		}
		privateRequest data.PrivateRequest
		response       data.Response
	)
	if err := bodyInto(&requestData); err != nil {
		return nil, werr(err, 400, `unable to understand request parameters`)
	}
	if err := requestData.PrivateRequest.Open(&privateRequest); err != nil {
		return nil, werr(err, 400, `unable to open private request envelope`)
	}
	if err := requestData.Data.Open(&response); err != nil {
		return nil, werr(err, 400, `unable to understand open response envelope`)
	} else if secret, err := privateRequest.Decode(response); err != nil {
		return nil, werr(err, 500, `unable to decrypt response`)
	} else {
		return textResponse(secret), nil
	}
}

func index(_ getBody) (response, *webError) {
	return htmlResponse(indexHTML), nil
}

func ListenAndServe(addr string) error {
	http.Handle(`/request`, handlerFunc(request))
	http.Handle(`/respond`, handlerFunc(respond))
	http.Handle(`/receive`, handlerFunc(receive))
	http.Handle(`/`, handlerFunc(index))
	err := http.ListenAndServe(addr, nil)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
