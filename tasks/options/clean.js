module.exports = function(config) {
  'use strict';

  return {
    release: ['<%= destDir %>', '<%= tempDir %>', '<%= genDir %>'],
    gen: ['<%= genDir %>'],
    temp: ['<%= tempDir %>'],
    css: ['<%= genDir %>/css'],
    vendor: ['<%= destDir %>/vendor'],
    build: ['<%= srcDir %>/app/*.js*','<%= srcDir %>/app/*.d.ts']
  };
};
