module.exports = function(config) {
  return {
    options: {
      encoding: 'utf8',
      algorithm: 'md5',
      length: 8,
    },
    js: {
      src: '<%= genDir %>/app/boot.js',
      dest: '<%= genDir %>/app'
    }
  };
};
