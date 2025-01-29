data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./services/agent/model",
    "--dialect", "mysql"
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  url = "mysql://mainuser:${getenv("MYSQL_PASSWORD")}@localhost:3306/codev"
  dev = "docker://mysql/8/dev"

  migration {
    dir = "file://services/agent/storage/migrations"
    format = golang-migrate
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}