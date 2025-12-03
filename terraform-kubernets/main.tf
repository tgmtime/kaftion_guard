# Terraform AWS EKS Modülü ile Kubernetes Kümesi
module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  version         = "~> 19.0"
  cluster_name    = "my-eks-cluster"
  cluster_version = "1.27"

  vpc_id     = var.vpc_id             # VPC ID'si
  subnets    = var.subnet_ids         # Alt ağların ID'leri

  node_groups = {
    eks_nodes = {
      desired_capacity = 2
      max_capacity     = 3
      min_capacity     = 1
      instance_type    = "t3.medium"
    }
  }
}

# EKS Cluster Yetkilendirme
data "aws_eks_cluster_auth" "auth" {
  name = module.eks.cluster_id
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
  token                  = data.aws_eks_cluster_auth.auth.token
}

provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
    token                  = data.aws_eks_cluster_auth.auth.token
  }
}

# Helm Chart ile Makefile Yükleme
resource "helm_release" "makefile_chart" {
  name       = "makefile-runner"
  chart      = "./helm/project-chart1" # Helm Chart yolu
  namespace  = "default"

  values = [
    file("./helm/project-chart1/values.yaml")
  ]

  set {
    name  = "makefile.hostPath"
    value = "/Users/enver/Archive/Study/Notes/Apache_Kafka/Makefile" # Makefile'ın yolu
  }

  set {
    name  = "makefile.target"
    value = "build" # Makefile komut hedefi
  }
}
