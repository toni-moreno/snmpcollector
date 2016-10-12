var gulp = require('gulp');
var todo = require('gulp-todo');
var clean = require('gulp-clean');


// generate a todo.md from your javascript files
gulp.task('todo', function() {
  gulp.src(['!Godeps/**','!node_modules/**','!vendor/**','**/*.js','**/*.go','**/*.sh'])
    .pipe(todo())
    .pipe(gulp.dest('./'));
  // -> Will output a TODO.md with  todos
});

gulp.task('clean-scripts', function () {
  return gulp.src(['!node_modules/**',
                    '!Godeps/**',
                    '!public/node_modules/**',
                    '!vendor/**',
                    '!public/systemjs.config.js',
                    'public/**/*.js',
                    'public/**/*.js.map'], {read: false})
    .pipe(clean());
});
