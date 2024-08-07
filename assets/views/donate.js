const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errorPickup: false,
    errorItems: false,
    errorCharity: false,
    hasError: false,
    donate: {
      resell: true,
      newItems: false,
      faith: true,
      itemTypes: [],
      charityTypes: [],
      anyCharityType: null,
      proximity: null,
      pickup: null,
      dropoff: null,
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

      this.hasError = false
      this.errorPickup = false
      this.errorItems = false
      this.errorCharity = false

      if (!this.donate.pickup && !this.donate.dropoff) {
        this.donate.pickup = true
        this.donate.dropoff = true
      }

      if (this.donate.itemTypes.length == 0) {
        this.errorItems = true
        this.hasError = true
      }

      if (this.donate.charityTypes.length == 0 && !this.donate.anyCharityType) {
        this.errorCharity = true
        this.hasError = true
      }

      if (this.hasError) {
        return
      }

      this.saveStep();
      window.location.assign("/donate/results")
    },
    reset(e) {
      e.preventDefault()

      this.hasError = false
      this.errorPickup = false
      this.errorItems = false
      this.errorCharity = false

      this.donate.itemTypes = []
      this.donate.charityTypes = []
      this.donate.resell = null
      this.donate.faith = null
      this.donate.newItems = null
      this.donate.anyCharityType = null
      this.donate.pickup = null
      this.donate.dropoff = null
      console.log('reset')
      localStorage.removeItem('donate')
    },
    saveStep() {
      localStorage.setItem('donate', JSON.stringify(this.donate))
    },
  }
})
