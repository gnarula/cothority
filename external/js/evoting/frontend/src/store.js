import Vue from 'vue'
import Vuex from 'vuex'
import createPersistedState from 'vuex-persistedstate'
import rosterTOML from './public.toml'
import cothority from '@dedis/cothority'

Vue.use(Vuex)

const net = cothority.net
const roster = cothority.Roster.fromTOML(rosterTOML)

console.log('Creating new store')

const store = new Vuex.Store({
  state: {
    user: null,
    loginReply: null,
    socket: new net.RosterSocket(roster, 'evoting')
  },
  getters: {
    isAuthenticated: state => {
      return state.user !== null
    },
    hasLoginReply: state => {
      return state.loginReply !== null
    }
  },
  mutations: {
    SET_LOGIN_REPLY (state, loginReply) {
      state.loginReply = loginReply
    },
    SET_USER (state, data) {
      state.user = data
    }
  },
  plugins: [createPersistedState({ key: 'evoting', paths: ['user'] })]
})

export default store
