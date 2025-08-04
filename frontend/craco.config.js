module.exports = {
  style: {
    postcss: {
      plugins: [
        require('tailwindcss'),
        require('autoprefixer'),
      ],
    },
  },
  jest: {
    configure: {
      transformIgnorePatterns: [
        'node_modules/(?!(axios)/)'
      ]
    }
  }
}