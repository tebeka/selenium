=================================
selenium - Selenium Client For Go
=================================

About
=====
This is a `Selenium`_ client for `Go`_.
Currently it supports only the remote WebDriver client, so you'll need a
`selenium server`_ running.

.. _`Selenium`: http://seleniumhq.org/
.. _`Go`: http://golang.org/
.. _`selenium server`: http://seleniumhq.org/download/

Installing
==========
Run
    `go get bitbucket.org/tebeka/selenium`

Docs
====
Docs are at gopkgdoc_. 

.. _gopkgdoc: http://gopkgdoc.appspot.com/pkg/bitbucket.org/tebeka/selenium



Changes
=======
See here_.

.. _here: ChangeLog

TODO
====
* Support Firefox profiles
* Finish full `Selenium API`_.
* More testing
* `Selenium 1`_ client
* Our own server for testing (getting out takes too much time)
* `SauceLabs`_ integration

.. _`Selenium API`: http://code.google.com/p/selenium/wiki/JsonWireProtocol
.. _`SauceLabs`: http://saucelabs.com/docs/quickstart
.. _`Selenium 1`: http://wiki.openqa.org/display/SRC/Specifications+for+Selenium+Remote+Control+Client+Driver+Protocol

Hacking
=======
This directory should be under an `src` directory.

You'll need a Selenium server to run the tests, run `selenium.sh download` to
get it and `selenium.sh start` to run it.

Test with `./run-tests.sh`.

I (Miki) work on `dev` branch since `go get` pull from default.

Authors
=======

* Miki Tebeka <miki.tebeka@gmail.com>

License
=======
`MIT`_

.. _MIT: https://bitbucket.org/tebeka/selenium/src/tip/LICENSE.txt
