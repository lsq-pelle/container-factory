package main

import (
	"archive/tar"
	"io"
	"log"
	"net/http"
	"path"

	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
)

var dock *docker.Client
var dockerAuth *docker.AuthConfigurations

func main() {
	var err error
	if dock, err = docker.NewClient("unix:///var/run/docker.sock"); err != nil || dock == nil {
		log.Fatal("couldn't initialise Docker", err)
	}

	if err := dock.Ping(); err != nil {
		log.Fatal("couldn't ping Docker", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/build", build).Methods("POST")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func build(res http.ResponseWriter, req *http.Request) {
	vars := req.URL.Query()

	authConfig, err := authFromHeaders(req.Header)
	if err != nil {
		panic(err)
	}

	buildStream := toReader(func(w io.Writer) error {
		return addBuildpack(w, req.Body)
	})

	bodyFormatter := toWriter(func(r io.Reader) error {
		return formatJSON(res, r)
	})

	buildOpts := docker.BuildImageOptions{
		Name:          vars.Get("t"),
		InputStream:   buildStream,
		OutputStream:  bodyFormatter,
		RawJSONStream: true,
		NoCache:       true,
	}

	if err := dock.BuildImage(buildOpts); err != nil {
		panic(err)
	}

	pushOpts := docker.PushImageOptions{
		Name:          buildOpts.Name,
		OutputStream:  bodyFormatter,
		RawJSONStream: true,
	}

	if err := dock.PushImage(pushOpts, authConfig); err != nil {
		panic(err)
	}
}

func addBuildpack(dest io.Writer, src io.ReadCloser) error {
	r := tar.NewReader(src)
	w := tar.NewWriter(dest)

	sawDockerfile := false
	filter := func(hdr *tar.Header) bool {
		if path.Clean(hdr.Name) == "Dockerfile" {
			sawDockerfile = true
		}
		return true
	}

	if err := copyTar(w, r, filter); err != nil {
		return err
	}

	if !sawDockerfile {
		if err := addFile(w, "packs/node/Dockerfile"); err != nil {
			return err
		}
	}

	return w.Flush()
}
