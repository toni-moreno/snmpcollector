module.exports = function(grunt) {
  "use strict";

  // Concat and Minify the src directory into dist
  grunt.registerTask('build', [
    'clean:build',
    'jshint:source',
    'jshint:tests',
    'jscs',
    'tslint',
    'clean:release',
   // 'copy:public_to_gen',
   // 'copy:node_modules',
    'typescript:build',
    // 'karma:test',
   // 'css',
    //'htmlmin:build',
    //'cssmin:build',
    //'ngAnnotate:build',
    //'systemjs:build',
   // 'concat:js',
   // 'filerev',
   // 'remapFilerev',
   // 'usemin',
   // 'uglify:genDir'
  ]);

  grunt.registerTask('build-post-process', function() {
    grunt.config('copy.public_gen_to_temp', {
      expand: true,
      cwd: '<%= genDir %>',
      src: '**/*',
      dest: '<%= tempDir %>/public/',
    });
    grunt.config('copy.backend_bin', {
      cwd: 'bin',
      expand: true,
      src: ['*'],
      options: { mode: true},
      dest: '<%= tempDir %>/bin/'
    });
    grunt.config('copy.backend_files', {
      expand: true,
      src: ['conf/defaults.ini', 'conf/sample.ini', 'vendor/**/*', 'scripts/*'],
      options: { mode: true},
      dest: '<%= tempDir %>'
    });

    grunt.task.run('copy:public_gen_to_temp');
    grunt.task.run('copy:backend_bin');
    grunt.task.run('copy:backend_files');
  });

};
