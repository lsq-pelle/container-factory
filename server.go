package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"github.com/nathan7/encoding-base32"
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

	r.HandleFunc("/{user}/{app}/{service}/push", build).Methods("POST")

	n := negroni.Classic()
	n.UseHandler(r)
	n.Run(":3000")
}

var encoding = base32.MinEncoding

func build(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	image := strings.Join([]string{vars["user"], vars["app"], vars["service"]}, "/")

	imageId, err := nameRegistry.Id(image)
	if err != nil {
		panic(err)
	}

	r, w := io.Pipe()
	go func() {
		w.CloseWithError(addBuildpack(w, req.Body))
	}()

	buildOpts := docker.BuildImageOptions{
		Name:          fmt.Sprintf("tutum.co/lsqio/%d", imageId),
		InputStream:   r,
		OutputStream:  os.Stdout,
		RawJSONStream: true,
		NoCache:       true,
	}

	if err := dock.BuildImage(buildOpts); err != nil {
		log.Fatal(err)
	}

	pushOpts := docker.PushImageOptions{
		Name:          buildOpts.Name,
		OutputStream:  os.Stdout,
		RawJSONStream: true,
		Registry:      "tutum.co",
	}

	if err := dock.PushImage(pushOpts, dockerAuth.Configs[pushOpts.Registry]); err != nil {
		log.Fatal(err)
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
