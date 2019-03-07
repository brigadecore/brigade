---
title: 'Tutorial 1: Writing a CI pipeline'
description: Writing your first CI pipeline, Part 1
section: intro
---

# Writing your first CI pipeline, Part 1

Let’s learn by example.

Throughout this tutorial, we’ll walk you through the creation of a basic web application with a Brigade CI pipeline for testing the application.

It’ll consist of two parts:

- A public site that lets people generate UUIDs.
- A brigade.js that tests the site

We’ll assume you have Brigade, git (a version control system), and pip (a package management system for Python) installed already.

You can tell Brigade is installed and which version by running the following command in a shell prompt (indicated by the $ prefix):

```
$ helm status brigade-server
```

If Brigade is installed, you should see the deployment status of your installation. If it isn’t, you’ll get an error telling "Error: getting deployed release "brigade-server": release: "brigade-server" not found".

See [Installing Brigade][install] for advice on how to install Brigade.

For pip, it is already installed if you're using Python 2 >=2.7.9 or Python 3 >=3.4 binaries downloaded from python.org.

You can tell pip is installed and which version by running the following command in a shell prompt:

```
$ pip --version
```

If you see "command not found", see [how to install pip](https://pip.pypa.io/en/stable/installing/) for advice on how to install pip.

If you're having trouble going through this tutorial, please post an issue to [Azure/brigade][github] to chat with other Brigade users who might be able to help.

## Creating your first application

For this tutorial, we'll be creating an example application written in Python which uses [Flask](http://flask.pocoo.org/) to provide a very simple UUID generator web server. Flask is a microframework for Python based on Werkzeug, Jinja 2 and good intentions.  (Note that this application is written for [Python 3](https://docs.python.org/3/), specifically.)

Let's write the web server. To create your app, type this command:

```
$ mkdir -p uuid-generator/app
```

Open the file `uuid-generator/app/__init__.py` and put the following Python code into it:

```python
from flask import Flask, Response
app = Flask(__name__)


@app.route("/")
def hello():
    return "Hello World!"
```

This is the simplest application possible in Flask. To call `hello()`, we need to map it to a URL - in this case we want to map it to the root path (`/`).

## Start the development server

Now that we've written the app, let's run it!

```
$ echo "Flask" > requirements.txt
$ pip install -r requirements.txt
$ FLASK_APP=uuid-generator/app/__init__.py flask run
 * Serving Flask app "app"
 * Running on http://127.0.0.1:5000/ (Press CTRL+C to quit)
```

Now, open a Web browser and go to “/” on your local domain – e.g., http://127.0.0.1:5000/. You should see "Hello World!":

![Hello World!](https://docs.brigade.sh/img/img1.png)

If you look back to the running flask app's logs, you should see new logs pop up:

```bash
127.0.0.1 - - [18/Aug/2017 12:56:16] "GET / HTTP/1.1" 200 -
```

## Generating UUIDs

Let's make the application generate a random UUID on every request. Edit the `uuid-generator/app/__init__.py` file so it looks like this:

```python
from flask import Flask, Response
import uuid

app = Flask(__name__)


@app.route("/")
def hello():
    return Response(str(uuid.uuid4()), status=200, mimetype='text/plain')
```

Re-run the web server and open the browser back to http://127.0.0.1:5000/. You should now see a random UUID:

!Random UUID](https://docs.brigade.sh/img/img2.png)

Keep refreshing the page. You should see new UUIDs being generated every time you refresh the page.

When you’re comfortable with the application, read [part 2 of this tutorial][part2] to learn about pushing our application to GitHub.

[github]: https://github.com/Azure/brigade
[install]: ../install
[part1]: ../tutorial01
[part2]: ../tutorial02