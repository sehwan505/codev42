data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./model",
    "--dialect", "mysql"
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  url = "mysql://mainuser:${getenv("MYSQL_PASSWORD")}@localhost:3306/codev?charset=utf8mb4&collation=utf8mb4_general_ci&tls=false"
  dev = "docker://maria/10.7/schema"

  migration {
    dir = "file://storage/migrations"
    format = golang-migrate
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}