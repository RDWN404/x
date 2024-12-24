pipeline {
    agent any

    stages {
        stage('Checkout') {
            steps {
                // Ambil kode dari Git repository
                echo "checking out repo"
                git url: 'https://github.com/RDWN404/x.git', branch: 'master'  // Ganti dengan URL repo Anda
            }
        }

        stage('Build Docker Image') {
            steps {
                script {
                    // Bangun Docker image
                    echo "starting docker build"
                    sh 'docker build -t ${DOCKER_IMAGE} .'
                    echo "docker built successfully"
                }
            }
        }

        // stage('Push Docker Image') {
        //     steps {
        //         echo "pushing to docker hub"
        //         script {
        //             // Push image ke Docker registry jika diperlukan
        //             // sh 'docker push ${DOCKER_IMAGE}'  // Un-comment jika ingin mem-push image
        //         }
        //         echo "done"
        //     }
        // }

        stage('Run Docker Container') {
            steps {
                script {
                    // Jalankan container dari Docker image
                    sh 'docker run -d -p 8080:8080 ${DOCKER_IMAGE}'
                }
            }
        }
    }

    post {
        always {
            // Cleanup: Hapus container setelah selesai
            sh 'docker ps -q | xargs docker stop | xargs docker rm'
        }
    }
}
