require("./spec_helper.js");

const migration = require("../201809181044_metron_agent_deployment_name.js");

describe("Metron Agent Deployment Name", function() {
  it("enables the cf metric name on upgrade", function(){
    migration.migrate(
      { properties: {} }
    ).should.deep.equal(
      { properties: { ".properties.enable_cf_metric_name": { "value": true } } }
    );
  });
});