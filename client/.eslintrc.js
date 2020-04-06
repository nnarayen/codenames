module.exports = {
  extends: ['airbnb-typescript'],
  parserOptions: {
    project: './tsconfig.json',
  },
  rules: {
    "jsx-a11y/click-events-have-key-events": "off",
    "react/no-array-index-key": "off",
  }
};
