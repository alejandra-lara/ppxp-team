require 'rspec'

class IntegrationSpecRunner
  class UnsupportedErtVersion < StandardError
  end

  SUPPORTED_ERT_VERSIONS = %w(1.5 1.6 1.7 1.8)

  def initialize(environment:, om_version:, ert_version:)
    fail 'No Environment Name provided' if environment.nil? || environment.empty?
    fail 'No Ops Manager Version provided' if om_version.nil? || om_version.empty?

    ENV['ENVIRONMENT_NAME'] = environment
    ENV['OM_VERSION'] = om_version

    unless SUPPORTED_ERT_VERSIONS.include?(ert_version)
      fail UnsupportedErtVersion, "Version #{ert_version.inspect} is not supported"
    end

    @ert_version = ert_version
  end

  def configure_ert
    run_spec(["integration/ERT-#{ert_version}/configure_ert_spec.rb"])
  end

  def configure_experimental_features
    run_spec(["integration/ERT-#{ert_version}/configure_experimental_features_spec.rb"])
  end

  def disable_http_traffic
    run_spec(["integration/ERT-#{ert_version}/disable_http_traffic.rb"])
  end

  def configure_external_dbs
    run_spec(["integration/ERT-#{ert_version}/configure_external_dbs_spec.rb"])
  end

  def configure_postgres
    unless ert_version == '1.6'
      fail UnsupportedErtVersion, "Version #{ert_version} is not supported for this task"
    end

    run_spec(["integration/ERT-#{ert_version}/configure_postgres_spec.rb"])
  end

  def configure_external_file_storage
    run_spec(["integration/ERT-#{ert_version}/configure_external_file_storage_spec.rb"])
  end

  def configure_ha_instance_counts
    run_spec(["integration/ERT-#{ert_version}/configure_ha_instance_counts_spec.rb"])
  end

  def configure_dea_instance_counts
    run_spec(["integration/ERT-#{ert_version}/configure_dea_instance_counts_spec.rb"])
  end

  def configure_aws_diego_cell_instance
    run_spec(["integration/ERT-#{ert_version}/configure_aws_diego_cell_instance_spec.rb"])
  end

  private

  def run_spec(spec_to_run)
    RSpecExiter.exit_rspec(RSpec::Core::Runner.run(spec_to_run))
  end

  attr_reader :ert_version
end

module RSpecExiter
  def self.exit_rspec(exit_code)
    exit exit_code
  end
end
