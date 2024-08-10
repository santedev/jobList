/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.html", "./**/*.go", "./**/*.templ"],
  safelist: [],
  theme: {
    extend: {
      screens: {
        '2xl': '1536px',
      },
    },
  },
}
