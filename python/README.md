# codeforge SDK for Python

The Python library to write [codeforge](https://codeforge.pointjungle.com) pipelines: it implements the run contract — reading the job's inputs, reporting results, sending logs, reading variables — so that a pipeline only writes its own logic.

Coming soon on PyPI as `pointjungle-codeforge-sdk` (package `codeforge_sdk`).

Examples will live in `examples/`: a pipeline run as a Python script and one run as a container, both started on your machine with `codeforge run --local`.
