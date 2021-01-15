@Library('dst-shared@master') _

dockerBuildPipeline {
        githubPushRepo = "Cray-HPE/hms-dns-dhcp"
        repository = "cray"
        imagePrefix = "hms"
        app = "dns-dhcp"
        name = "hms-dns-dhcp"
        description = "Cray HMS common DNS and DHCP interfaces."
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "internal"
}

