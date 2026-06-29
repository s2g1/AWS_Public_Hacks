/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    screens: {
      'sm': '768px',   // tablet breakpoint (768-1023px)
      'lg': '1024px',  // desktop breakpoint (1024px+)
    },
    extend: {},
  },
  plugins: [],
}
