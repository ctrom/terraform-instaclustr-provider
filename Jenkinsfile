
String repositoryName = env.JOB_NAME.split('/')[0]

if(env.BRANCH_NAME == 'master') {

    properties([buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '7', numToKeepStr: '10'))])

    node ("generic-build-agent") {
        checkout scm

        ws("${env.JENKINS_HOME}/workspace/go/src/github.com/PeopleNet/${repositoryName}") {

            
            sh "mkdir -p ${env.JENKINS_HOME}/workspace/go/src/github.com/PeopleNet/${repositoryName}"
            checkout(
                    [$class                           : 'GitSCM', branches: [[name: "*/${env.BRANCH_NAME}"]],
                    doGenerateSubmoduleConfigurations: false,
                    extensions                       : [[$class: 'WipeWorkspace']],
                    submoduleCfg                     : [], userRemoteConfigs:
                            [[credentialsId: 'peoplenet-ci2',
                            url          : "git@github.com:PeopleNet/${repositoryName}.git"]]]
            )
            withEnv([
                    "GOPATH=/var/jenkins_home/workspace/go"
            ]) {

                tool('go')
                sh """
                git describe --tags > version
                VERSION=\$(cat version)
                echo \$VERSION
                CGO_ENABLED=0 GOOS=linux /usr/local/go/bin/go  build -installsuffix cgo -o terraform-provider-instaclustr_v\$VERSION
                zip terraform-provider-instaclustr-\$VERSION-linux-amd64.zip terraform-provider-instaclustr_v\$VERSION
                rm terraform-provider-instaclustr_v\$VERSION
                GOOS=darwin /usr/local/go/bin/go  build -o terraform-provider-instaclustr_v\$VERSION
                zip terraform-provider-instaclustr-\$VERSION-darwin-amd64.zip terraform-provider-instaclustr_v\$VERSION
                """

                tool 'aws_cli'

                withCredentials([
                        [$class: 'UsernamePasswordMultiBinding', credentialsId: 'aws', usernameVariable: 'AWS_ACCESS_KEY_ID', passwordVariable: 'AWS_SECRET_ACCESS_KEY'],
                ]) {
                    sh """
                    git describe --tags > version
                    VERSION=\$(cat version)
                    aws s3 cp terraform-provider-instaclustr-\$VERSION-linux-amd64.zip s3://peoplenet-custom-tools/terraform-provider-instaclustr/terraform-provider-instaclustr-\$VERSION-linux-amd64.zip
                    aws s3 cp terraform-provider-instaclustr-\$VERSION-darwin-amd64.zip s3://peoplenet-custom-tools/terraform-provider-instaclustr/terraform-provider-instaclustr-\$VERSION-darwin-amd64.zip
                    """
                }
            }



            deleteDir()
        }
    }

}

