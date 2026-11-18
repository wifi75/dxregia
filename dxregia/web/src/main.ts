import './glass.css'
import './controls.css'
import './theme.css'
import './theme.svelte'
import { mount } from 'svelte'
import App from './App.svelte'

mount(App, { target: document.getElementById('app')! })
