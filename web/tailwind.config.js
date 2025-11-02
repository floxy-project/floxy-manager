/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    './public/index.html',
    './src/**/*.{ts,tsx,js,jsx,html}'
  ],
  theme: {
    extend: {
      colors: {
        // Neutral surfaces tuned for glassmorphism
        surface: {
          DEFAULT: 'rgba(255,255,255,0.6)',
          dark: 'rgba(17,24,39,0.6)'
        },
        // Orange/Gold colors for dark theme
        'orange-dark': {
          50: '#fff5ed',
          100: '#ffe0d1',
          200: '#ffc3a3',
          300: '#ff9d6b',
          400: '#ff6b35',
          500: '#ff4500',
          600: '#e63e00',
          700: '#cc3700',
          800: '#b33000',
          900: '#9a2900'
        },
        'gold': {
          50: '#fffef7',
          100: '#fffce5',
          200: '#fff8b3',
          300: '#fff280',
          400: '#ffed4d',
          500: '#d4af37',
          600: '#c19f2f',
          700: '#ae8f27',
          800: '#9b7f1f',
          900: '#886f17'
        },
        // VS Code/Cursor style dark colors
        'ide-dark': {
          bg: '#1e1e1e',
          surface: '#252526',
          border: '#3e3e42',
          text: '#ff6b35',
          'text-secondary': '#ff8c42',
          'text-muted': '#ffa366'
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      boxShadow: {
        glass: '0 8px 32px rgba(0,0,0,0.12)'
      }
    }
  },
  plugins: []
};
