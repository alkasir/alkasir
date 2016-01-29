/* global __dirname, process, module, require */
var webpack = require("webpack");

var entry = {
    app: ["./browser/script/index"]
};

// Add react hot reloader if ALKASIR_HOT is set
if (process.env.ALKASIR_HOT !== undefined ) {
    entry.app = [
        "webpack-dev-server/client?http://0.0.0.0:3000", // WebpackDevServer host and port
        "webpack/hot/only-dev-server",
        "./browser/script/index"
    ];
}

var c = {
    module: {
        loaders: [
            { test: /\.js$/, exclude: /node_modules/, loader: "babel-loader"},
            { test: /\.jsx$/, loaders: ["react-hot-loader", "babel-loader"] },
            { test: /\.coffee$/, loader: "coffee" },
            { test: /\.less$/, loader: "style!css!less" },
            { test: /\.css$/, loader: "style!css" },
            { test: /\.woff$/, loader: "file" },
            { test: /\.woff2$/, loader: "file" },
            { test: /\.ttf$/, loader: "file" },
            { test: /\.eot$/, loader: "file" },
            { test: /\.svg$/, loader: "file" },
            { test: /res\/messages\/.*\.json$/, loaders: ["json-loader"], exclude: /node_modules/ },
            { test: /res\/documents\/.*\.md$/, loaders: ["raw-loader"], exclude: /node_modules/ }
        ]
    },
    plugins: [
        new webpack.HotModuleReplacementPlugin()
        // new webpack.NoErrorsPlugin()
    ],
    entry: entry,
    output: {
        path: __dirname + "/res/generated/",
        filename: "[name].bundle.js",
        publicPath: "/generated/"
    },
    resolve: {
        extensions: ["", ".web.coffee", ".web.js", ".coffee", ".js", ".jsx"]
    },
    devtool: "cheap-source-map"
};


module.exports = c;
