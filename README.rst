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
You'll need a Selenium server to run the tests, run `selenium.sh download` to
get it and `selenium.sh start` to run it.

Test with `gotest -v`.

I (Miki) work on `dev` branch since `goinstall` pull from default.

Authors
=======

* Miki Tebeka <miki.tebeka@gmail.com>


License
=======
Copyright (C) 2010 Miki Tebeka <miki.tebeka@gmail.com>

Distributed under the Eclipse Public License, the same as Clojure.
