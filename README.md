# Golibs

Golibs is a collection of independent go packages. It provides reusable components that you can decide to use on your projects.
This is not a framework. You can pick the parts you want and decide to not use the rest.
It is 100% compatible with the standard library and does not try to reinvent its own custom interfaces.

No premature abstraction! Nothing should be added to this library unless it is used on at least 3 live projects.

# Dependencies

The library assume that logrus is used as a logger.
Some package might require specific dependencies.

# Libraries

## Http

The http package provides useful components to build a web API (commonly used middlewares, standard responses, etc.)

## Metrics

The metrics package is a wrapper around common metric technologies. Currently only datadog is supported.
It provides a logrus-like interface to work with tags.

## Queue

The queue package provides the basic tools to build a worker. It sends messages via a go channel and hides all the polling logic.
