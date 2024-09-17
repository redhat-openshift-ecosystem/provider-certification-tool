module "workers_instances" {
  count = 2
  source = "../ec2_instance"
  instance_type = var.instance_type
  subnet_id = var.subnet_id
  cluster_member_sg_id =var.cluster_member_sg_id
  cluster_instance_profile_name = var.cluster_instance_profile_name
  ami_id = var.ami_id
  key_name = var.key_name
}
