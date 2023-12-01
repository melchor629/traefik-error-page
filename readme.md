# traefik error page

An adaptation of the [error](https://github.com/traefik/traefik/blob/master/pkg/middlewares/customerrors/custom_errors.go) middleware from traefik that works similar but allows to respond from the service only if the response from the real service has no body.

> **NOTE**: this code contains code from the aformentioned middleware.

## Configuration

Works similar as in the middleware, but there are extra configuration and the `service` works differently.

```yaml
http:
  middlewares:
    error-pages:
      plugin:
        traefik-error-page:
          # range of status codes that the middleware will accept
          status: ['400-499', '500-599']
          # this should point to the root url of the service, the middleware cannot access to certain
          # information from traefik like services, so this should point to the scheme://host[:port]
          service: 'https://http.cat'
          # the path to make the request to
          query: '/{status}'
          # only replace the response with the error from the other service if the real service
          # answered without any content (true by default)
          emptyOnly: true
          # prints some logs
          debug: false
```
