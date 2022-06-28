import adapter from '@sveltejs/adapter-node';
import preprocess from 'svelte-preprocess';
import { optimizeImports } from 'carbon-preprocess-svelte';

const counters = {};

const increment = (id) => {
  if (!counters[id]) {
    counters[id] = 0;
  }

  return counters[id]++;
};

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://github.com/sveltejs/svelte-preprocess
  // for more information about preprocessors
  preprocess: [preprocess(), optimizeImports()],

  kit: {
    adapter: adapter(),
  },

  compilerOptions: {
    cssHash: ({ hash, css, name, filename }) => `${name}-${increment(name)}`,
  },
};

export default config;
