# kubecat
> Minimal network uptime monitor for in-cluster kubernetes and docker cluster services. Also can be run stand-alone for general uptime checks.

![](./screenshot.png)


Reporting services:
- [Sentry.io](https://sentry.io)

Reporters:
- HTTP `modules.HTTP`
- Tile38 `modules.Tile38`


Example `config.yaml`

```yaml
reporters:
  - name: api-service
    module: "http"
    interval: 60 # time in seconds
    options:
      method: POST
      url: http://localhost:3000/v1/users
      timeout: 30000
      acceptableStatus:
        - 200
        - 201
      headers:
        jwtToken: "jwt token here"

  - name: tile38-object-check
    module: "Tile38"
    interval: 120 # time in seconds
    options:
      url: http://tile38:9851
      timeout: 30 # in seconds
      min: 20 // minimum num_objects in the tile38 database
```

Run:
`SENTRY_DSN=<keyhere> ./kubecat`

In docker:
`docker run -e SENTRY_DSN=<keyhere> -v ./config.yaml:/app/config.yaml stevelacy/kubecat`


### [MIT](./LICENSE)
