exports.migrate = function(input) {

  input.variable_migrations.push({
    from: input.properties['.mysql.autoscale_credentials'],
    to_variable: 'autoscale-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.ccdb_credentials'],
    to_variable: 'cc-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.diegodb_credentials'],
    to_variable: 'diego-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.locketdb_credentials'],
    to_variable: 'locket-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.mysql_backup_server_credentials'],
    to_variable: 'mysql-backup-server-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.mysql_bootstrap_credentials'],
    to_variable: 'mysql-bootstrap-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.diag_agent_credentials'],
    to_variable: 'mysql-diag-agent-db-credentials'
  });

  if (input.properties['.mysql.mysqlmetricsdb_credentials'] != null) {
    input.variable_migrations.push({
      from: input.properties['.mysql.mysqlmetricsdb_credentials'],
      to_variable: 'mysql-metrics-db-credentials'
    });
  }

  input.variable_migrations.push({
    from: input.properties['.mysql.monitordb_credentials'],
    to_variable: 'mysql-monitor-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.networkpolicyserverdb_credentials'],
    to_variable: 'network-policy-server-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.nfsbrokerpush.nfs_broker_push_credentials'],
    to_variable: 'nfs-broker-push-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.nfsvolume_credentials'],
    to_variable: 'nfs-volume-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.notifications_credentials'],
    to_variable: 'notifications-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.pivotal_account_credentials'],
    to_variable: 'pivotal-account-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.routingdb_credentials'],
    to_variable: 'routing-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.silkdb_credentials'],
    to_variable: 'silk-db-credentials'
  });

  input.variable_migrations.push({
    from: input.properties['.mysql.uaadb_credentials'],
    to_variable: 'uaa-db-credentials'
  });

  return input;
};
