package main

import (
	"fmt"
	"io"
	"log"
	"regexp"

	"github.com/deis/quokka/pkg/javascript/libk8s"
	"gopkg.in/gin-gonic/gin.v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/acid/pkg/config"
	"github.com/deis/acid/pkg/storage"
	"github.com/deis/acid/pkg/webhook"
)

func logToHTML(c *gin.Context) {
	namespace, _ := config.AcidNamespace(c)
	org := c.Param("org")
	proj := c.Param("project")
	commit := c.Param("commit")

	if proj == "favicon.ico" {
		c.JSON(404, gin.H{"message": "Not found"})
		return
	}

	pname := fmt.Sprintf("%s/%s", org, proj)
	log.Printf("Loading logs for %q", pname)
	p, err := storage.New().Get(pname, namespace)
	if err != nil {
		log.Printf("logToHTML: error loading project: %s", err)
	}

	c.Writer.Write([]byte(bootstrapHead))
	defer c.Writer.Write([]byte(bootstrapFoot))

	if len(commit) > 0 {
		if !SHAish(commit) {
			last, err := webhook.GetLastCommit(p, commit)
			if err != nil {
				log.Printf("error parsing commit reference %s: %s", commit, err)
				c.Writer.WriteString("No reference")
				return
			}
			commit = last
		}
	}
	fmt.Fprintf(c.Writer, "<p>Logs for Git reference %q</p>", commit)

	name := fmt.Sprintf("github.com-%s-%s", org, proj)
	pods, err := taskPods(commit, name, namespace)
	if err != nil {
		log.Printf("could not load task pods: %s", err)
		c.Writer.WriteString("No task pods found. Has the job started?")
		return
	}

	panelHead := `<div class="panel panel-%s"><div class="panel-heading"><h3 class="panel-title">%s</h3></div><div class="panel-body"><pre>`
	panelFoot := `</pre></div></div>`
	for _, p := range pods.Items {
		st := "info"
		switch p.Status.Phase {
		case v1.PodPending, v1.PodRunning:
			st = "warning"
		case v1.PodFailed:
			st = "danger"
		case v1.PodSucceeded:
			st = "success"

		}

		// Print out logs for this item
		fmt.Fprintf(c.Writer, panelHead, st, p.Labels["jobname"])
		podLog(p.ObjectMeta.Name, namespace, c.Writer)
		fmt.Fprintln(c.Writer, panelFoot)
	}
}

func podLog(name, namespace string, w io.Writer) error {
	kc, err := libk8s.KubeClient()
	if err != nil {
		return err
	}

	req := kc.CoreV1().Pods(namespace).GetLogs(name, &v1.PodLogOptions{})

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	_, err = io.Copy(w, readCloser)
	return err
}

// taskPods gets the pods associated with this task
func taskPods(commit, name, namespace string) (*v1.PodList, error) {
	// Load the pods that ran as part of this build.
	kc, err := libk8s.KubeClient()
	if err != nil {
		return nil, err
	}

	lo := v1.ListOptions{LabelSelector: fmt.Sprintf("commit=%s,belongsto=%s", commit, name)}

	return kc.CoreV1().Pods(namespace).List(lo)
}

var sha = regexp.MustCompile("^[a-f0-9]{40}$")

// SHAish returns true if the given string looks like a GitHub SHA
func SHAish(s string) bool {
	return sha.MatchString(s)
}

func badge(c *gin.Context) {
	org := c.Param("org")
	proj := c.Param("project")
	namespace, _ := config.AcidNamespace(c)

	c.Writer.Header().Set("content-type", "image/svg+xml;charset=utf-8")

	pname := fmt.Sprintf("%s/%s", org, proj)
	log.Printf("Loading project %s", pname)
	p, err := storage.New().Get(pname, namespace)
	if err != nil {
		log.Printf("badge: error loading project: %s", err)
		c.Writer.WriteString(badgeFailing)
		return
	}

	status, err := webhook.GetRepoStatus(p, "master")
	if err != nil {
		log.Printf("badge: error fetching status: %s", err)
		c.Writer.WriteString(badgeFailing)
		return
	}

	badge := badgeRunning
	switch *status.State {
	case webhook.StateSuccess:
		badge = badgePassing
	case webhook.StatusFailure, webhook.StatusError:
		log.Printf("badge: marked build failed because status was %s - %q", *status.State, *status.Description)
		badge = badgeFailing
	default:
		log.Printf("badge: status was %s - %q", *status.State, *status.Description)
	}
	c.Writer.WriteString(badge)
}

const bootstrapHead = `

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <!-- The above 3 meta tags *must* come first in the head; any other head content must come *after* these tags -->
    <meta name="description" content="">
    <meta name="author" content="">
    <link rel="icon" href="../../favicon.ico">

    <title>Acid Logs</title>

	<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">

<!-- Optional theme -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">

    <!-- HTML5 shim and Respond.js for IE8 support of HTML5 elements and media queries -->
    <!--[if lt IE 9]>
      <script src="https://oss.maxcdn.com/html5shiv/3.7.3/html5shiv.min.js"></script>
      <script src="https://oss.maxcdn.com/respond/1.4.2/respond.min.js"></script>
    <![endif]-->

	<style>
	body {
  padding-top: 50px;
}
.starter-template {
  padding: 40px 15px;
}
</style>
  </head>

  <body>

    <nav class="navbar navbar-inverse navbar-fixed-top">
      <div class="container">
        <div class="navbar-header">
          <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#navbar" aria-expanded="false" aria-controls="navbar">
            <span class="sr-only">Toggle navigation</span>
            <span class="icon-bar"></span>
            <span class="icon-bar"></span>
            <span class="icon-bar"></span>
          </button>
          <a class="navbar-brand" href="#">Acid</a>
        </div>
        <div id="navbar" class="collapse navbar-collapse">
          <ul class="nav navbar-nav">
            <li class="active"><a href="#">Home</a></li>
            <li><a href="#about">About</a></li>
            <li><a href="#contact">Contact</a></li>
          </ul>
        </div><!--/.nav-collapse -->
      </div>
    </nav>

    <div class="container">
      <div class="starter-template">
        <h1>Log Output</h1>
`
const bootstrapFoot = `
      </div>
    </div><!-- /.container -->


    <!-- Bootstrap core JavaScript
    ================================================== -->
    <!-- Placed at the end of the document so the pages load faster -->
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js"></script>
    <script>window.jQuery || document.write('<script src="../../assets/js/vendor/jquery.min.js"><\/script>')</script>

<!-- Latest compiled and minified JavaScript -->
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>

  </body>
</html>
`

// The following SVGs were all generated by http://shields.io/#your-badge

const badgePassing = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="68" height="20"><linearGradient id="b" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="a"><rect width="68" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#a)"><path fill="#555" d="M0 0h33v20H0z"/><path fill="#4c1" d="M33 0h35v20H33z"/><path fill="url(#b)" d="M0 0h68v20H0z"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="16.5" y="15" fill="#010101" fill-opacity=".3">acid</text><text x="16.5" y="14">acid</text><text x="49.5" y="15" fill="#010101" fill-opacity=".3">pass</text><text x="49.5" y="14">pass</text></g></svg>`
const badgeFailing = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="60" height="20"><linearGradient id="b" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="a"><rect width="60" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#a)"><path fill="#555" d="M0 0h33v20H0z"/><path fill="#e05d44" d="M33 0h27v20H33z"/><path fill="url(#b)" d="M0 0h60v20H0z"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="16.5" y="15" fill="#010101" fill-opacity=".3">acid</text><text x="16.5" y="14">acid</text><text x="45.5" y="15" fill="#010101" fill-opacity=".3">fail</text><text x="45.5" y="14">fail</text></g></svg>`
const badgeRunning = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="86" height="20"><linearGradient id="b" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="a"><rect width="86" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#a)"><path fill="#555" d="M0 0h33v20H0z"/><path fill="#dfb317" d="M33 0h53v20H33z"/><path fill="url(#b)" d="M0 0h86v20H0z"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="16.5" y="15" fill="#010101" fill-opacity=".3">acid</text><text x="16.5" y="14">acid</text><text x="58.5" y="15" fill="#010101" fill-opacity=".3">running</text><text x="58.5" y="14">running</text></g></svg>`
