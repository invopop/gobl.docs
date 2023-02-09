require "json"
require "active_support/inflector"

# Automate the generation of Mintlify's navigation json

ActiveSupport::Inflector.inflections do |inflect|
  inflect.acronym "GOBL"
  inflect.acronym "UUID"
  inflect.acronym "URL"
  inflect.acronym "DSig"
  inflect.acronym "ID"
  inflect.acronym "CBC"
end

DOCS_DIR = "draft-0"

pages = Hash.new { |hash, key| hash[key] = [] }

Dir
  .glob("#{DOCS_DIR}/**/*.mdx")
  .each do |file|
    dir = File.dirname(file).split("/").last
    path = dir == DOCS_DIR ? dir : "#{DOCS_DIR}/" + dir
    page = File.basename(file, ".mdx")
    pages[dir] << "#{path}/#{page}"
  end

parent_pages = pages[DOCS_DIR]
pages.delete(DOCS_DIR)

puts JSON.generate(
       {
         group: DOCS_DIR,
         pages: [
           parent_pages,
           pages.map { |k, v| { group: k.camelize, pages: v } },
         ].flatten,
       },
     )
