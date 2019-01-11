# vcs-torture - Version control torture tests

The vcs-torture project aims to answer this question: what are the
limits to specific version control systems?

How well do version controls systems perform in general? What happens when
your repo is large? What is large for a specific version control system? How
long do typical operations take as the limits are approached? What are the
practical limits?

Use this program to get answers for specific version control systems on
specific hardware. This is not a "which version control system is best"
benchmark. Instead, it aims to point out the practical limits for usage for
many version control systems.

The term "version control" refers to the use programs like Git, Mercurial,
Subversion, Perforce, CVS, RCS, SCCS, and so on, in putting many versions of
files into some "repository", and getting specific versions of files out. This
is synonymous with the term "source control", and is a subset of another field
called "configuration management".

## Getting started

Getting started is as easy as having a Go compiler installed on
your system (Go 1.11 or higher, because the code uses modules), and:

```
$ git clone git@github.com:neurocline/vcs-torture.git
$ cd vcs-torture
$ go build .
$ ./vcs-torture [options...]
```

Many of the tests take a long time and use a lot of disk space. The larger
tests are impractical without an SSD (it can take weeks to build up a repo
large enough to stress some of the limits of version control systems).

## Currently supported

The following version control systems are supported

- Git
- Mercurial

The following limits are tested

- number of files in worktree
- size of repository: bytes, objects, commits
- operations: init, add/commit, branch, checkout, clone

## What's next?

Add more version control systems. Here's the planned order

- Subversion
- Perforce

If there are other version control systems you want to see put to the
test, then ask. Or, better, submit pull requests. For practical reasons,
it has to be something that most people can run; submitting a pull
request for an expensive or private version control system isn't likely
to be accepted.

Add more kinds of tests.

- network
- history
- merging

## About the code

This code uses Go modules in a perhaps idiosyncratic way. The code is not designed
to be sharable, but putting code in packages for namespacing is a good thing.

If you look at the `go.mod` file in the root of the project, you'll see it
names the top-level package as simply `vcs-torture`. Since no other Go code
can include this top-level package (it's `package main`), all this does is
let the rest of the code include packages as `vcs-torture/<package>`, instead
of the much longer full-URL style.

This is completely Go 1.11 compliant; it's a side benefit of not needing $GOPATH
any more.
