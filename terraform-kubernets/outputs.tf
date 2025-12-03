output "cluster_name" {
  value       = module.eks.cluster_id
  description = "EKS kümesinin adı"
}

output "kubeconfig" {
  value       = module.eks.kubeconfig
  sensitive   = true
  description = "Kubeconfig dosyası"
}

output "helm_status" {
  value       = helm_release.makefile_chart.status
  description = "Helm release durumu"
}
