# -*- encoding: utf-8 -*-
lib = File.expand_path('../docs/lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'drycc-docs/version'

Gem::Specification.new do |gem|
  gem.name          = "drycc-docs"
  gem.version       = DryccDocs::VERSION
  gem.authors       = ["Jesse Stuart"]
  gem.email         = ["jesse@jessestuart.ca"]
  gem.description   = %q{Drycc docs}
  gem.summary       = %q{Drycc docs}
  gem.homepage      = ""

  gem.files         = `git ls-files docs`.split($/)
  gem.executables   = gem.files.grep(%r{^bin/}).map{ |f| File.basename(f) }
  gem.test_files    = gem.files.grep(%r{^(test|spec|features)/})
  gem.require_paths = ["docs/lib"]
end
