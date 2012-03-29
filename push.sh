#!/bin/bash
# Push both to bitbucket and github

hg push
hg push git+ssh://git@github.com/tebeka/selenium.git
