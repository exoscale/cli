{
  default-service-strategy: allow,
  services: {
    iam: {
      type: allow
    }
  }
}
 run . iam org-policy update -
