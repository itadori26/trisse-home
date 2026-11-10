import { mount } from 'svelte';
import App from './app/App.svelte';
import './shared/theme/tokens.css';
import './shared/theme/base.css';

const target = document.getElementById('app');
if (!target) throw new Error('missing #app element');

export default mount(App, { target });
