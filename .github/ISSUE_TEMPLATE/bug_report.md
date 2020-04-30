---
name: Bug report
about: Create a report to help us improve

---

### Prerequisites

* [ ] I am running the [latest version](https://github.com/dan-v/awslambdaproxy/releases) of awslambdaproxy
* [ ] I am have read the [README](https://github.com/dan-v/awslambdaproxy#usage) instructions and the [FAQ](https://github.com/dan-v/awslambdaproxy#faq)

### Description

[Description of the issue]

### Steps to Reproduce

1. [First Step]
2. [Second Step]
3. [and so on...]

**Expected behavior:** [What you expected to happen]

**Actual behavior:** [What actually happened]

### Environment
* If you are using CLI, get the version and specify the full command you are using.
```
./awslambdaproxy version
awslambdaproxy version 0.0.12
./awslambdaproxy -r us-west-2,us-west-1 -f 60
```
* If you are using Docker, get the version and specify the full command you are using.
```
docker run -it --rm --entrypoint /app/awslambdaproxy vdan/awslambdaproxy -v
awslambdaproxy version 0.0.12
docker run -d vdan/awslambdaproxy -r us-west-2,us-west-1 -f 60
```

### Error Output
```
...
```