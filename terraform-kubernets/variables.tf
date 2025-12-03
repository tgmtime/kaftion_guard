variable "aws_region" {
  description = "AWS bölgesi"
  type        = string
  default     = "us-east-1"
}

variable "vpc_id" {
  description = "EKS kümesinin çalışacağı VPC ID'si"
  type        = string
}

variable "subnet_ids" {
  description = "EKS kümesi için kullanılacak alt ağların ID'leri"
  type        = list(string)
}
