const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errors: [],
    donate: {
      itemTypes: [],
      charityTypes: [],
      proximity: null,
      pickupDropoff: null,
      zip: null,
    }
  },
  mounted() {
    try {
      donate = JSON.parse(localStorage.getItem('donate'));
      if (donate != null) {
        this.donate = donate
      }
    } catch(e) {
      localStorage.removeItem('donate')
    }
  },
  methods:{
    checkForm: function (e) {
      e.preventDefault();
      this.errors = [];
      this.saveStep();
      window.location.assign("/donate/step/2")
    },
    saveStep() {
      localStorage.setItem('donate', JSON.stringify(this.donate))
    },
  }
})
