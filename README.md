# JSONWatcher

**this service not yet completed and is still in its early stages. Please don't consider it as a production release.**

A service to watch `json` file changes and view it on stdout using `less`.

As I'm a developer, many times I log objects on terminal for inspection during development. I found that it takes approx. 30 seconds to copy dumped json from terminal to IDE, format it, and then insepct it. I decided to develop a service that I can run on my second screen and see the dumped json updated while I'm working, therefor saving my time.

## How to use

1. build the service
    ```sh
    GOOS=linux go build -o jsonwatcher ./...
    ```

run the service and provide a file to watch
```sh
./jsonwatcher -f myfile.json
```

to exit press `q`, this will close `less` and take you to the file changes log. Press `q` again to terminate the program.