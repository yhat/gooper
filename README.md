# gooper your Goopfiles

Wait. What my what?

`gooper` is a simple Go dependency manager for Github. It allows you to peg
packages to specific commits, rather than always pulling from HEAD, and
uncover your own project's dependencies.

It uses Goopfiles inspired by [goop](https://github.com/nitrous-io/goop).

## What a Goopfile looks like:

Goopfiles contain links to Github packages with corresponding git commit SHAs.

    github.com/yhat/go-project      329971d209e4d517218a2c2a171c96d4692dcb07
    github.com/yhat/another-project
    github.com/yhat/a-third-project d66284b75e77744a31341d0bd47c390c16fbc0bf

## Install gooper

    go get github.com/yhat/gooper

## Commands

List all github depencencies of your current project with the current SHA.

    gooper freeze [optional: path to go files]

Install all packages in your Gooperfile and revert to the pegged commit.

    gooper install [optional: path to Goopfile]
