const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errorPickup: false,
    errorItems: false,
    errorCharity: false,
    hasError: false,
    donate: {
      itemTypes: [],
      charityTypes: [],
      anyCharityType: null,
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
      console.log(this.donate.anyCharityType)

      this.hasError = false
      this.errorPickup = false
      this.errorItems = false
      this.errorCharity = false

      if (!this.donate.pickupDropoff) {
        this.errorPickup = true
        this.hasError = true
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
    saveStep() {
      localStorage.setItem('donate', JSON.stringify(this.donate))
    },
  }
})
