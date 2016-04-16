module.exports = function(config) {
  return {
    // copy source to temp, we will minify in place for the dist build
    everything_but_less_to_temp: {
      cwd: '<%= srcDir %>',
      expand: true,
      src: ['**/*', '!**/*.less'],
      dest: '<%= tempDir %>'
    },

    public_to_gen: {
      cwd: '<%= srcDir %>',
      expand: true,
      src: ['**/*', '!**/*.less'],
      dest: '<%= genDir %>'
    },

    node_modules: {
      cwd: './node_modules',
      expand: true,
      src: [
        'angular2/bundles/*',
	'angular2/platform/*',
        'angular2/*.d.ts',
        'angular2/typings/**/*',
        'angular2/manual_typings/**/*',
	'angular2/es6/dev/src/testing/*',
        'systemjs/dist/*',
        'es6-promise/**/*',
        'es5-shim/*.js',
        'es6-shim/*',
        'reflect-metadata/*',
        'rxjs/bundles/*',
        'rxjs/Rx.d.ts',
      ],
      dest: '<%= srcDir %>/vendor/npm'
    }

  };
};
