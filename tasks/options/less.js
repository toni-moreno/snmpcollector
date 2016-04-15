module.exports = function(config) {

  return {
    src:{
      options: {
        paths: ["<%= srcDir %>/vendor/bootstrap/less", "<%= srcDir %>/less"],
        yuicompress: true
      },
      files: {
        "<%= genDir %>/css/bootstrap-responsive.min.css": "<%= srcDir %>/less/influxsnmp-responsive.less"
      }
    }
  };
};
