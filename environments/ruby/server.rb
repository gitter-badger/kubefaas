# frozen_string_literal: true

require 'rack'
require 'thin'
require 'logger'

require_relative 'kubefaas/specializer'
require_relative 'kubefaas/handler'

$handler = nil

app = Rack::Builder.new do
  use Rack::Logger, Logger::DEBUG
  use Rack::CommonLogger

  map "/specialize" do
    run Kubefaas::Specializer
  end

  map '/v2/specialize' do
    run Kubefaas::V2::Specializer
  end

  map "/healthz" do
    run ->(env) { [ 200, {}, [] ] }
  end

  map '/' do
    run Kubefaas::Handler
  end
end

Rack::Handler::Thin.run app, Host: '0.0.0.0', Port: 8888
