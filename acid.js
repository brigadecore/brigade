pr = pushRecord;

// This is a Go project, so we want to set it up for Go.
gopath = "/go"

// To set GOPATH correctly, we have to override the default
// path that Acid sets.
localPath = gopath + "/src/github.com/" + pr.repository.name

job1 = {
  name: "run-unit-tests",
  image: "acid-go:latest",
  env: {
    "DEST_PATH": localPath,
    "GOPATH": gopath
  },
  tasks:[
    "go get github.com/Masterminds/glide",
    "glide install",
    "make test-fast"
  ]
}

run(job1, pr)
