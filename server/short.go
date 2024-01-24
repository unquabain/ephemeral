package server

import (
	"html/template"
	"io"
	"net/http"

	// embed needs to be imported to enable the go:embed special compiler comment.
	_ "embed"
	"net/url"

	"github.com/Unquabain/ephemeral/data"
	"github.com/Unquabain/ephemeral/envelope"
	"github.com/apex/log"
)

type shortPhase int

//go:embed shortRequest.html
var shortRequestHTML string

//go:embed shortRespond.html
var shortRespondHTML string

//go:embed error.html
var errorHTML string

const (
	requestPhase shortPhase = iota
	respondGetPhase
	respondPostPhase
	receivePhase
)

func detectPhase(r *http.Request) (shortPhase, map[string]string) {
	public := r.FormValue(`public`)
	private := r.FormValue(`private`)
	data := r.FormValue(`data`)
	if r.Method == http.MethodGet {
		if public == `` {
			return requestPhase, nil
		}
		return respondGetPhase, map[string]string{`public`: public}
	}
	if public != `` && data != `` {
		return respondPostPhase, map[string]string{`public`: public, `data`: data}
	}
	if private != `` && data != `` {
		return receivePhase, map[string]string{`private`: private, `data`: data}
	}
	return requestPhase, nil
}

func shortError(w http.ResponseWriter, r *http.Request, err error, msg string) {
	log.
		WithError(err).
		WithField(`url`, r.URL.String()).
		Error(msg)
	w.Header().Add(`Content-Type`, `text/html`)
	tmplt := template.Must(template.New(`error`).Parse(errorHTML))
	if err := tmplt.Execute(w, struct{ Msg string }{msg}); err != nil {
		log.WithError(err).Error(`could not render error page`)
	}
}

func returnURL(request *http.Request) string {
	var ret url.URL
	ret.Scheme = `https`
	ret.Host = request.Host
	ret.User = request.URL.User
	ret.Path = request.URL.Path
	log.WithFields(log.Fields{
		`request url`: request.URL,
		`Scheme`:      `https`,
		`Host`:        request.Host,
		`User`:        request.URL.User,
		`Path`:        request.URL.Path,
		`return url`:  ret.String(),
	}).Debug(`return URL`)
	return ret.String()
}

func shortRequest(w http.ResponseWriter, r *http.Request) {
	var tctx struct {
		Public, Private, URL string
	}
	tctx.URL = returnURL(r)
	if private, err := data.NewPrivateKey(data.RandomCurve()); err != nil {
		shortError(w, r, err, `could not create private key`)
		return
	} else if public, err := private.Public().MarshalText(); err != nil {
		shortError(w, r, err, `could not marshal public key`)
		return
	} else if private, err := private.MarshalText(); err != nil {
		shortError(w, r, err, `could not marshal private key`)
		return
	} else if tmplt, err := template.New(`request`).Parse(shortRequestHTML); err != nil {
		shortError(w, r, err, `could not parse template`)
		return
	} else {
		tctx.Public = string(public)
		tctx.Private = string(private)
		if err := tmplt.Execute(w, tctx); err != nil {
			shortError(w, r, err, `could not render template`)
		}
	}
}

func shortRespondGet(w http.ResponseWriter, r *http.Request, dict map[string]string) {
	if t, err := template.New(`respond`).Parse(shortRespondHTML); err != nil {
		shortError(w, r, err, `could not parse template`)
		return
	} else if err := t.Execute(w, dict); err != nil {
		shortError(w, r, err, `could not render template`)
	}
}

func shortRespondPost(w http.ResponseWriter, r *http.Request, dict map[string]string) {
	var request data.PublicRequest
	var env envelope.Envelope
	if err := request.Key.UnmarshalText([]byte(dict[`public`])); err != nil {
		shortError(w, r, err, `could not parse public key`)
		return
	}
	response, err := request.Encode([]byte(dict[`data`]))
	if err != nil {
		shortError(w, r, err, `could not encode data`)
		return
	}
	env.Prelude = `Send this back to the person who sent you this link.`
	env.Name = `RESPONSE`
	if err := env.Stuff(response); err != nil {
		shortError(w, r, err, `could not stuff response envelope`)
		return
	}
	w.Header().Add(`Content-Type`, `text/plain`)
	if _, err := io.Copy(w, env.Reader()); err != nil {
		shortError(w, r, err, `could not write envelope`)
		return
	}
}
func shortReceive(w http.ResponseWriter, r *http.Request, dict map[string]string) {
	var request data.PrivateRequest
	var env envelope.Envelope
	var response data.Response
	if err := request.Key.UnmarshalText([]byte(dict[`private`])); err != nil {
		shortError(w, r, err, `could not parse private key`)
		return
	}

	if err := env.UnmarshalText([]byte(dict[`data`])); err != nil {
		shortError(w, r, err, `could not read envelope`)
		return
	}

	err := env.Open(&response)
	if err != nil {
		shortError(w, r, err, `could not open envelope`)
		return
	}

	secret, err := request.Decode(response)
	if err != nil {
		shortError(w, r, err, `could not decode secret`)
		return
	}

	w.Header().Add(`Content-Type`, `text/plain`)
	w.Header().Add(`Content-Disposition`, `attachment; filename="secret.txt"`)
	if _, err := w.Write(secret); err != nil {
		shortError(w, r, err, `could not write secret`)
		return
	}
}
func short(w http.ResponseWriter, r *http.Request) {
	phase, dict := detectPhase(r)
	switch phase {
	case requestPhase:
		shortRequest(w, r)
	case respondGetPhase:
		shortRespondGet(w, r, dict)
	case respondPostPhase:
		shortRespondPost(w, r, dict)
	case receivePhase:
		shortReceive(w, r, dict)
	}

}
