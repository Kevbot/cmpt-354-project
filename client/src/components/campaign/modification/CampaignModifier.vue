<!-- CampaignModifier.vue -->

<template>
  <v-container>
    <v-row>
      <!-- Campaign selection dropdown -->
      <v-col cols="4">
        <p class="headline primary--text">Select your campaign</p>
        <v-select
          solo
          no-data-text="You haven't created any campaigns yet."
          prepend-inner-icon="mdi-account-search"
          :items="campaignDropdownItems"
          item-text="name"
          item-value="id"
          v-model="selectedCampaignID"
          @change="fetchAuxiliaryCampaignData"
        >
          <!-- Button used to refresh the campaign list. -->
          <template v-slot:append-outer>
            <v-btn
              color="primary"
              fab
              small
              class="mt-n2"
              @click="fetchPlayerCreatedCampaigns"
            >
              <v-icon>mdi-refresh</v-icon>
            </v-btn>
          </template>
        </v-select>
      </v-col>
    </v-row>

    <!-- Campaign Name -->
    <v-card v-if="campaigns.length" class="pl-2">
      <v-card-title class="display-1 primary--text">
        {{ selectedCampaign.name }}
      </v-card-title>

      <!-- Buttons -->
      <v-card-actions>
        <v-btn class="primary" @click="displayCampaignEditorPrompt">
          Edit Campaign
          <v-icon class="ml-2">mdi-pencil</v-icon>
        </v-btn>

        <v-btn class="error" @click="displayCampaignDeletionPrompt">
          Delete Campaign
          <v-icon>mdi-delete-forever</v-icon>
        </v-btn>
      </v-card-actions>

      <ModifyCampaignForm
        ref="modifyCampaign"
        @complete="sendCampaignUpdateRequest"
      >
      </ModifyCampaignForm>

      <ConfirmDialog ref="confirmDelete" />

      <!-- This container displays all campaign information -->
      <v-container>
        <p class="headline primary--text">Campaign information</p>
        <v-row>
          <v-col cols="12" md="8" lg="8">
            <v-text-field
              readonly
              label="Location"
              :value="selectedCampaign.current_location"
            ></v-text-field>
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="12" lg="12">
            <v-textarea
              readonly
              no-resize
              label="State"
              outlined
              :value="selectedCampaign.state"
            ></v-textarea>
          </v-col>
        </v-row>

        <!-- Display some data bout the campaign participants. -->
        <p class="headline primary--text">Campaign participants</p>
        <v-row class="mb-4">
          <v-simple-table>
            <thead>
              <tr>
                <th class="primary--text subtitle-1">Player</th>
                <th class="primary--text subtitle-1">Character</th>
                <th class="primary--text subtitle-1">Race</th>
                <th class="primary--text subtitle-1">Class</th>
                <th class="primary--text subtitle-1">
                  Attendance record for your campaigns
                </th>
              </tr>
            </thead>

            <tbody>
              <tr v-if="participants.length === 0">
                <td>This campaign has no participants!</td>
              </tr>

              <tr
                v-else
                v-for="(participant, index) in participants"
                :key="index"
              >
                <td class="subtitle-1">
                  {{ participant.player_username }}
                </td>

                <td class="subtitle-1">
                  {{ participant.character_name }}
                </td>

                <td class="subtitle-1">
                  {{ participant.character_race }}
                </td>

                <td class="subtitle-1">
                  {{ participant.character_class }}
                </td>

                <td class="subtitle-1">
                  {{ getPlayerAttendanceStatus(participant.player_username) }}
                </td>
              </tr>
            </tbody>
          </v-simple-table>
        </v-row>

        <!-- Campaign milestone information -->
        <p class="headline primary--text">Campaign milestones</p>
        <v-btn @click="displayAddMilestoneForm">
          Add Milestone
          <v-icon class="mx-2">mdi-plus</v-icon>
        </v-btn>

        <AddMilestoneForm
          ref="addMilestone"
          @complete="sendAddMilestoneRequest"
        >
        </AddMilestoneForm>

        <v-row>
          <v-col cols="12" lg="12">
            <MilestoneTimeline :milestones="milestones"></MilestoneTimeline>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-container>
</template>

<script>
import { mapActions } from 'vuex';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import AddMilestoneForm from './AddMilestoneForm.vue';
import MilestoneTimeline from './MilestoneTimeline.vue';
import ModifyCampaignForm from './ModifyCampaignForm.vue';

export default {
  name: 'CampaignModifier',
  components: {
    ConfirmDialog,
    AddMilestoneForm,
    MilestoneTimeline,
    ModifyCampaignForm,
  },
  data() {
    return {
      campaigns: [],
      selectedCampaignID: null,
      milestones: [],
      participants: [],
      perfectPlayers: [],
    };
  },
  /**
   * Fetches all of the user's character and campaign data when the component loads.
   */
  async created() {
    await this.fetchPlayerCreatedCampaigns();

    if (this.selectedCampaignID === null) return;
    await this.fetchAuxiliaryCampaignData();
  },
  computed: {
    /**
     * @returns {Array<Object>} An array of objects which contain the ID
     * and name of each campaign.
     */
    campaignDropdownItems() {
      if (this.campaigns === null) return [];
      return this.campaigns.map(({ id, name }) => ({ id, name }));
    },
    /**
     * @returns {Number} The campaign currently selected by
     * the user.
     */
    selectedCampaign() {
      if (!this.selectedCampaignID) return {};

      const target = this.selectedCampaignID;
      return this.campaigns.filter((x) => x.id === target)[0];
    },
  },
  methods: {
    ...mapActions({
      display: 'notifications/display',
    }),
    /**
     * Attempts to fetch all data that requires the campaign to first be
     * selected.
     */
    async fetchAuxiliaryCampaignData() {
      await this.fetchCampaignMilestones();
      await this.fetchCampaignParticipants();
      await this.fetchPlayersWithPerfectAttendance();
    },
    /**
     * Attempts to retrieve all campaigns which have been created by the
     * logged in player.
     */
    async fetchPlayerCreatedCampaigns() {
      const requestURI = 'auth/campaign/me';
      const method = 'GET';

      await this.$http({
        url: requestURI,
        data: null,
        method,
      })
        .then((resp) => {
          const fetchedCampaigns = resp.data.data.campaigns;
          if (fetchedCampaigns) {
            this.campaigns = fetchedCampaigns;
            this.selectedCampaignID = fetchedCampaigns[0].id;
          } else {
            this.campaigns = [];
            this.selectedCampaignID = null;
          }
        })
        .catch((err) => {
          let message = 'Could not fetch campaigns. Please try again.';
          if (err.response) {
            message = `Error: ${err.response.data.message}. Please try again.`;
          }
          this.display({
            message,
            color: 'error',
            timeout: 10000,
          });
        });
    },
    /**
     * Shows the user a form which will allow them to add milestones for
     * a campaign.
     */
    async displayAddMilestoneForm() {
      await this.$refs.addMilestone.prompt();
    },
    /**
     * Send an HTTP request to create a milestone for the given campaign.
     * @param {String} milestone - The milestone text.
     */
    async sendAddMilestoneRequest({ milestone }) {
      const campaignID = parseInt(this.selectedCampaignID, 10);
      const requestURI = 'auth/campaign/milestone';
      const method = 'POST';

      await this.$http({
        url: requestURI,
        data: {
          campaign_id: campaignID,
          milestone,
        },
        method,
      })
        .then(() => {
          this.display({
            message: 'Milestone was successfully created!',
            color: 'success',
            timeout: 6000,
          });
        })
        .catch((err) => {
          let message = 'Something went wrong. Please try again.';
          if (err.response) {
            message = `Error: ${err.response.data.message}. Please try again.`;
          }
          this.display({
            message,
            color: 'error',
            timeout: 6000,
          });
        })
        .finally(() => {
          this.fetchCampaignMilestones();
        });
    },
    /**
     * @returns {Array<String>} An array of milestones for the currently-selected campaign.
     */
    async fetchCampaignMilestones() {
      const campaignID = parseInt(this.selectedCampaignID, 10);
      const requestURI = `auth/campaign/${campaignID}/milestone`;
      const method = 'GET';
      await this.$http({
        url: requestURI,
        data: null,
        method,
      })
        .then((resp) => {
          const fetchedMilestones = resp.data.data.milestones;
          this.milestones = fetchedMilestones ?? [];
        })
        .catch((err) => {
          let message = 'Something went wrong. Please try again.';
          if (err.response) {
            message = `Error: ${err.response.data.message}. Please try again.`;
          }
          this.display({
            message,
            color: 'error',
            timeout: 6000,
          });
        });
    },
    /**
     * Displays a form to the user which allows them to input a new
     * campaign state and location.
     */
    async displayCampaignEditorPrompt() {
      const { state, current_location: location } = this.selectedCampaign;
      await this.$refs.modifyCampaign.prompt(state, location);
    },
    /**
     * Send a request to update the currently-selected campaign's state
     * and location.
     */
    async sendCampaignUpdateRequest({ state, location }) {
      const integerID = parseInt(this.selectedCampaignID, 10);

      const requestURI = `auth/campaign/${integerID}`;
      const method = 'PUT';

      await this.$http({
        url: requestURI,
        data: {
          state,
          location,
        },
        method,
      })
        .then(() => {
          this.display({
            message: 'Campaign was successfully updated!',
            color: 'success',
            timeout: 6000,
          });
        })
        .catch((err) => {
          let message = 'Something went wrong. Please try again.';
          if (err.response) {
            message = `Error: ${err.response.data.message}. Please try again.`;
          }
          this.display({
            message,
            color: 'error',
            timeout: 6000,
          });
        })
        .finally(async () => {
          /* Reset the current campaign to the campaign that was just
          updated rather than back to index 0 of the fetched campaigns. */
          const resetID = this.selectedCampaignID;
          await this.fetchPlayerCreatedCampaigns();
          this.selectedCampaignID = resetID;
        });
    },
    /**
     * Displays a prompt to the user, asking them to confirm their
     * campaign deletion.
     */
    async displayCampaignDeletionPrompt() {
      const title = 'Confirm deletion';
      const message = 'Delete your campaign? This action is irreversible.';
      if (await this.$refs.confirmDelete.prompt(title, message)) {
        await this.requestCampaignDeletion();
      }
    },
    /**
     * Sends an HTTP request to delete the campaign selected by the
     * user.
     */
    async requestCampaignDeletion() {
      const integerID = parseInt(this.selectedCampaignID, 10);

      const requestURI = `auth/campaign/${integerID}`;
      const method = 'DELETE';

      await this.$http({
        url: requestURI,
        data: null,
        method,
      })
        .then(() => {
          this.display({
            message: 'Campaign was successfully deleted!',
            color: 'success',
            timeout: 6000,
          });
        })
        .catch((err) => {
          let message = 'Something went wrong. Please try again.';
          if (err.response) {
            message = `Error: ${err.response.data.message}. Please try again.`;
          }
          this.display({
            message,
            color: 'error',
            timeout: 6000,
          });
        })
        .finally(() => {
          this.fetchPlayerCreatedCampaigns();
        });
    },
    /**
     * Retrieves some data about the players and characters
     * participating in the selected campaign.
     */
    async fetchCampaignParticipants() {
      const campaignID = parseInt(this.selectedCampaignID, 10);
      const requestURI = `auth/campaign/${campaignID}/participants`;
      const method = 'GET';
      await this.$http({
        url: requestURI,
        data: null,
        method,
      })
        .then((resp) => {
          const fetchedparticipants = resp.data.data.participants;
          this.participants = fetchedparticipants ?? [];
        })
        .catch((err) => {
          let message = 'Something went wrong. Please try again.';
          if (err.response) {
            message = `Error: ${err.response.data.message}. Please try again.`;
          }
          this.display({
            message,
            color: 'error',
            timeout: 6000,
          });
        });
    },
    /**
     * Fetch all players who have a perfect participation record for
     * campaigns created by the logged-in user.
     */
    async fetchPlayersWithPerfectAttendance() {
      const requestURI = '/auth/campaign/me/stats/player-attendance';
      const method = 'GET';
      await this.$http({
        url: requestURI,
        data: null,
        method,
      })
        .then((resp) => {
          const fetchedUsernames = resp.data.data.usernames;
          this.perfectPlayers = fetchedUsernames ?? [];
        })
        .catch((err) => {
          let message = 'Something went wrong. Please try again.';
          if (err.response) {
            message = `Error: ${err.response.data.message}. Please try again.`;
          }
          this.display({
            message,
            color: 'error',
            timeout: 6000,
          });
        });
    },
    /**
     * @param {String} username - The username of the player to check.
     * @returns {String} - A text description of the player's attendance
     * record. It's all-or-nothing at the moment.
     */
    getPlayerAttendanceStatus(username) {
      if (!this.perfectPlayers) return 'Not perfec attendance';
      const isPerfect = this.perfectPlayers.includes(username);
      return isPerfect ? 'Perfect attendance' : 'Not perfect attendance';
    },
  },
};
</script>
