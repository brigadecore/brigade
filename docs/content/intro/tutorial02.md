---
title: 'Tutorial 2: Setup GitHub'
description: Writing your first CI pipeline, Part 2
section: intro
---

# Writing your first CI pipeline, Part 2

This tutorial begins where [Tutorial 1][part1] left off. We’ll walk through the process for using Git locally on your personal computer, and using Github to back it up. We'll walk through creating your personal Github account, setting up Git on your computer, starting your first Git repository, and connecting that repository to a Github repository.

## What are git and Github?

Git is a widely-used version control system used to manage code. Git allows you to save drafts of your code so that you can look back at previous versions and potentially undo complicated errors. A project managed with git is called a *git repository*.

Github is a popular hosting service for git repositories. Github allows you to store your local Git repositories in the cloud. With Github, you can backup your personal files, share your code, and collaborate with others through their site.

## Create a Github account

To use Github, you will need a Github account.

In your own browser:

1. Open a new browser tab
2. Navigate to https://github.com/
3. Create an account

If you already have Github account, continue to the next exercise.

After you sign up, you will receive a verification e-mail. Be sure to verify your e-mail address to Github by following the instructions in that e-mail.

## Initialize your repository

Now that you have a Github account and a working application, let's create a git repository and push it up to Github!

Let's first change working directories over to the Flask app we made in [Tutorial 1][part1].

```
$ cd uuid-generator
```

Now, we'll create a git repository and make your first commit.

```
$ git init
$ echo -e "*.pyc\n*.egg-info" > .gitignore
$ git commit -am "initial commit"
```

We created a *.gitignore* file so all compiled python files and python eggs (noted with the .pyc and .egg-info extensions, respectively) are not added to source control.

## Push to Github

Finally, we'll create a repository on Github and then link it to your local repository on your computer. This will allow you to back up your work safely so you never need to worry about losing your work again!

1. On Github, create a new repository by navigating to https://github.com/new
2. On the new repository page, give your repository a name. It's not necessary, but it would be convenient to name it the same as the directory, **uuid-generator**
3. Add a nice description, leave it as a public repository and do not check the box to initialize this repository with a README
4. Click *Create repository*
5. After creating a repository, Github displays the repository page. The repository is empty, so it's time to connect it to your existing work. Copy the git commands on the Github page, under the title "...or push an existing repository from the command line", and paste them into your terminal. Running these commands will add a remote repository, and then push your local repository to the remote repository
6. Once your terminal reports that the push is complete, refresh the page on Github. You should now see the app uploaded to Github!

Github automatically displays the contents of a file named README if it exists in the repository. The README file is the perfect place to write a description of your project, if you so choose.

There you have it! Your first Github repository, linked to your local git repository.

When you’re comfortable with the git workflow, read [part 3 of this tutorial][part3] to learn about configuring our Github repository to test new features using Brigade.

[part1]: ../tutorial01
[part3]: ../tutorial03