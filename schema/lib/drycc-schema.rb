require 'drycc-schema/version'

module DryccSchema
  def self.dir
    @dir ||= File.expand_path(File.join(__FILE__, '..', '..'))
  end

  def self.paths
    @paths ||= Dir[File.join(dir, '**', '*.json')].reject do |path|
      path.index(File.join(dir, 'lib', ''))
    end
  end
end
