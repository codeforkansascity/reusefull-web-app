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

      if (!this.donate.pickup) {
        this.errors.push('Pickup or drop-off is required. ')
      }
      if (!this.donate.proximity) {
        this.errors.push('Proximity to charity is required. ')
      }

      if (this.errors.length > 0 ) {
        return;
      }

      this.saveStep();
      window.location.assign("/donate/search")
    },
    saveStep() {
      localStorage.setItem('donate', JSON.stringify(this.donate))
    },
  }
})
