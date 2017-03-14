# AppEngine Flexible Environment Custom Runtime Example

This is an example AppEngine app that includes a Docker container definition
that provides all of the dependencies needed to run Selenium WebDriver with
Firefox on AppEngine.

This is only provided as an example for now. It will not currently be supported
as a base image for other Dockerfiles.

This Dockerfile was cobbled together by looking at the [Java8 custom
image](https://github.com/GoogleCloudPlatform/openjdk-runtime/blob/master/openjdk8/src/main/docker/Dockerfile)
provided by Google for AppEngine. It can easily be made better; consider this a
draft.

## Installation

1. [Install gcloud and create an AppEngine
project](https://cloud.google.com/appengine/docs/flexible/go/quickstart).
2. Run the following:

   ```
   $ go get github.com/tebeka/selenium google.golang.org/appengine/cmd/aedeploy
   $ cd $GOROOT/src/github.com/tebeka/selenium/vendor
   $ go run init.go
   $ cd $GOROOT/src/github.com/tebeka/selenium/appengine
   $ cp ../vendor/selenium-server-standalone-3.3.1.jar ../vendor/geckodriver-v0.15.0-linux64 .
   $ aedeploy gcloud app deploy
   ... takes a while ...
   ```
3. Visit https://<your project ID>.appspot.com. A screenshot of google.com will
   be shown, which was obtained by running Firefox through WebDriver.
