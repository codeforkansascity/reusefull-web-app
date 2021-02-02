const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errors: [],
    donate: {
      itemTypes: [],
      charityTypes: [],
      budget: null,
      proximity: null,
      pickupDropoff: null,
      zip: null
    },
    charities: []
  },
  created() {
    try {
      donate = JSON.parse(localStorage.getItem('donate'));
      if (donate != null) {
        this.donate = donate
        fetch('/api/v1/donate/search', {
          method: 'post',
          body: JSON.stringify(this.donate)
        }).then(response => {
          response.json().then(charities => {
            this.charities = charities
          })
        })
      }
    } catch(e) {
      localStorage.removeItem('donate')
    }
  },
  methods:{
    search: function (e) {
      e.preventDefault();

      fetch('/api/v1/donate/search', {
        method: 'post',
        body: JSON.stringify(this.donate)
      }).then(response => {
        console.log(response)
        // if (response.ok) {
        //   window.location.assign("/charity/signup/thankyou");
        // } else {
        //   this.processing = false
        //   if (response.status == 409) {
        //     this.serverError = 'That email has already been registered.'
        //   } else {
        //     this.serverError= 'Error registering. Please try again later.'
        //   }
        //   return false
        // }
      })
    },
  }
})
