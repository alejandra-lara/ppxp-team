resource "aws_db_instance" "rds" {
  allocated_storage    = 100
  instance_class       = "${var.rds_instance_class}"
  engine               = "mysql"
  engine_version       = "5.6.22"
  name                 = "${var.rds_db_name}"
  username             = "${var.rds_db_username}"
  password             = "${var.rds_db_password}"
  db_subnet_group_name = "${aws_db_subnet_group.rds_subnet_group.name}"
  publicly_accessible = false
  vpc_security_group_ids = ["${aws_security_group.mysql_security_group.id}"]
  iops = 1000
  multi_az = true

  count = "${var.rds_instance_count}"
}

output "rds_address" {
  value = "${aws_db_instance.rds.address}"
}

output "rds_port" {
  value = "${aws_db_instance.rds.port}"
}

output "rds_username" {
  value = "${aws_db_instance.rds.username}"
}

output "rds_password" {
  value = "${var.rds_db_password}"
}

output "rds_db_name" {
  value = "${aws_db_instance.rds.name}"
}
