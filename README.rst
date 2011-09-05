=================================
selenium - Selenium Client For Go
=================================

About
=====
This is a `Selenium`_ client for `Go`_.
Currently it supports on the remote client, so you'll need a selenium server
running.

.. _`Selenium`: http://seleniumhq.org/
.. _`Go`: http://golang.org/


Authors
=======

* Miki Tebeka <miki.tebeka@gmail.com>

Changes
=======
See here_.

.. _here: https://bitbucket.org/tebeka/go-selenium/src/tip/ChangeLog

TODO
====
* Support screenshots
* Support Firefox profiles
* Finish full `Selenium API`_.
* More testing
* Our own server for testing (getting out takes too much time)
* `SauceLabs`_ integration

.. _`Selenium API`: http://code.google.com/p/selenium/wiki/JsonWireProtocol
.. _`SauceLabs`: http://saucelabs.com/docs/quickstart

Hacking
=======
You'll need a Selenium server to run the tests, run `selenium.sh download` to
get it and `selenium.sh start` to run it.

Test with `gotest -v`.

License
=======
Copyright (C) 2010 Miki Tebeka <miki.tebeka@gmail.com>

Distributed under the Eclipse Public License, the same as Clojure.
