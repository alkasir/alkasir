/* global process */
module.exports = function(grunt) {
    grunt.initConfig({
        webstore_upload: {
            "accounts": {
                "default": {
                    publish: true,
                    client_id: process.env.CLIENT_ID,
                    client_secret: process.env.CLIENT_SECRET
                },
            },
            "extensions": {
                "extension1": {
                    appID: process.env.APP_ID,
                    zip: "../src/src.zip"

                }
            }
        }

    });
    grunt.loadNpmTasks('grunt-webstore-upload');
    grunt.registerTask('default', ['webstore_upload']);
};
