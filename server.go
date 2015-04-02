package main

import (
	"archive/tar"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/negroni"
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

	if dockerAuth, err = docker.NewAuthConfigurationsFromDockerCfg(); err != nil || dockerAuth == nil {
		log.Fatal("couldn't initialise Docker auth", err)
	}

	if err := dock.Ping(); err != nil {
		log.Fatal("couldn't ping Docker", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/", build).Methods("POST")

	n := negroni.Classic()
	n.UseHandler(r)
	n.Run(":3000")
}

func build(res http.ResponseWriter, req *http.Request) {
	r, w := io.Pipe()
	go func() {
		w.CloseWithError(addBuildpack(w, req.Body))
	}()

	buildOpts := docker.BuildImageOptions{
		Name:         "tutum.co/lsqio/app",
		InputStream:  r,
		OutputStream: os.Stdout,
		NoCache:      true,
	}

	if err := dock.BuildImage(buildOpts); err != nil {
		log.Fatal(err)
	}

	pushOpts := docker.PushImageOptions{
		Name:         buildOpts.Name,
		OutputStream: os.Stdout,
		Registry:     "tutum.co",
	}

	if err := dock.PushImage(pushOpts, dockerAuth.Configs[pushOpts.Registry]); err != nil {
		log.Fatal(err)
	}
}

func addBuildpack(dest io.Writer, src io.ReadCloser) error {
	r := tar.NewReader(src)
	w := tar.NewWriter(dest)

	if err := copyTar(w, r); err != nil {
		return err
	}

	if err := addFile(w, "packs/node/Dockerfile"); err != nil {
		return err
	}

	return w.Flush()
}
