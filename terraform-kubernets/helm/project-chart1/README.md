Adım 1: Helm Chart'ı oluşturun

 helm create project-chart1


Adım 3: Helm Chart'ı yükleyin
 helm install my-makefile ./project-chart1


Adım 4: Pod'un durumunu kontrol edin
 kubectl get pods
 kubectl logs <pod-name>
