@Library('jenkins-pipeline') _

node {
  cleanWs()

  repo = "exoscale/cli"

  try {
    dir('src') {
      stage('SCM') {
        checkout scm
      }
      updateGithubCommitStatus('PENDING', "${env.WORKSPACE}/src")
      stage('Build') {
        parallel (
          "gofmt": {
            gofmt()
          },
          "go lint": {
            golint(repo)
          },
          "go test": {
            test()
          },
          "go build": {
            build()
          }
        )
      }
    }
  } catch (err) {
    currentBuild.result = 'FAILURE'
    throw err
  } finally {
    if (!currentBuild.result) {
      currentBuild.result = 'SUCCESS'
    }
    updateGithubCommitStatus(currentBuild.result, "${env.WORKSPACE}/src")
    cleanWs cleanWhenFailure: false
  }
}

def gofmt() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.pull()
    image.inside("-u root --net=host") {
      sh 'test $(gofmt -s -d -e $(find -iname "*.go" | grep -v "/vendor/") | tee -a /dev/fd/2 | wc -l) -eq 0'
    }
  }
}

def golint(repo) {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.pull()
    image.inside("-u root --net=host -v ${env.WORKSPACE}/src:/go/src/github.com/${repo}") {
      sh "go vet `go list github.com/${repo}/... | grep -v /vendor/`"
      sh "cd /go/src/github.com/${repo} && golangci-lint run ./..."
    }
  }
}

def test() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.inside("-u root --net=host") {
      sh "go test -v -mod=vendor ./..."
    }
  }
}

def build() {
  docker.withRegistry('https://registry.internal.exoscale.ch') {
    def image = docker.image('registry.internal.exoscale.ch/exoscale/golang:1.11')
    image.inside("-u root --net=host") {
      sh "go build -mod vendor -o exo"
      sh "test -e exo"
    }
  }
}
