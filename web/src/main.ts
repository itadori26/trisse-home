import { mount } from 'svelte';
import './app/featureMessages';
import App from './app/App.svelte';
import { registerServiceWorker } from './shared/pwa';
import './shared/theme/tokens.css';
import './shared/theme/base.css';

const target = document.getElementById('app');
if (!target) throw new Error('missing #app element');

registerServiceWorker();

export default mount(App, { target });
