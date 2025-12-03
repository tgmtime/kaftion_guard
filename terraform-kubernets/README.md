1. Terraform'u Başlatma
 terraform init

2. Planlama
 terraform plan -var 'vpc_id=<your-vpc-id>' -var 'subnet_ids=["<subnet-id-1>", "<subnet-id-2>"]'


3.Kaynakları Oluşturma

 terraform apply -var 'vpc_id=<your-vpc-id>' -var 'subnet_ids=["<subnet-id-1>", "<subnet-id-2>"]'



4.Helm Chart'ın Kurulumunu Kontrol Etme 

 kubectl get pods
 kubectl logs <pod-name>
