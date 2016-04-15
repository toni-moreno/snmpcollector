module.exports = function() {
  'use strict';

  return {
    html: [
      '<%= genDir %>/views/index.html',
      '<%= genDir %>/views/500.html',
    ],
    options: {
      assetsDirs: ['<%= genDir %>'],
      patterns: {
        css: [
          [/(\.css)/, 'Replacing reference to image.png']
        ]
      }
    }
  };
};
