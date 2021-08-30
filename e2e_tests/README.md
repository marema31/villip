# End to end tests for villip

To start the test :

```
docker compose up --build
````

The test output will be available in outputs folder

If you need to rerun the test without recompiling the villip binary (because you have added a test in the testsuite for example)
```
docker compose start venom
```