# gooper your Goopfiles

## Install

    go get github.com/yhat/gooper

## Commands

List the current dependencies with:

    gooper freeze [optional: path to go files]

Install from a Goopfile with:

    gooper install [optional: path to Goopfile]

## What a Goopfile looks like:

Goopfiles contain links to github projects and a corresponding git SHA

    github.com/yhat/go-project      329971d209e4d517218a2c2a171c96d4692dcb07
    github.com/yhat/another-project
    github.com/yhat/a-third-project d66284b75e77744a31341d0bd47c390c16fbc0bf
