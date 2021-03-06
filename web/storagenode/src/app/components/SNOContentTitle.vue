// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

<template>
    <div class="title-area">
        <h1 class="title-area__title">Your Storage Node Stats</h1>
        <div class="title-area__info-container">
            <div class="title-area__info-container__info-item">
                <p class="title-area__info-container__info-item__title">STATUS</p>
                <p v-if="online" class="title-area__info-container__info-item__content online-status">Online</p>
                <p v-else class="title-area__info-container__info-item__content offline-status">Offline</p>
            </div>
            <div class="title-area-divider"></div>
            <div class="title-area__info-container__info-item">
                <p class="title-area__info-container__info-item__title">UPTIME</p>
                <P class="title-area__info-container__info-item__content">{{uptime}}</P>
            </div>
            <div class="title-area-divider"></div>
            <div class="title-area__info-container__info-item">
                <p class="title-area__info-container__info-item__title">LAST CONTACT</p>
                <P class="title-area__info-container__info-item__content">{{lastPinged}} ago</P>
            </div>
            <div class="title-area-divider"></div>
            <VInfo
                v-if="info.isLastVersion"
                text="Running the minimal allowed version:"
                :bold-text="info.allowedVersion"
            >
                <div class="title-area__info-container__info-item">
                    <p class="title-area__info-container__info-item__title">VERSION</p>
                    <P class="title-area__info-container__info-item__content">{{info.version}}</P>
                </div>
            </VInfo>
            <VInfo
                v-if="!info.isLastVersion"
                text="Your node is outdated. Please update to:"
                bold-text="v0.0.0"
            >
                <div class="title-area__info-container__info-item">
                    <p class="title-area__info-container__info-item__title">VERSION</p>
                    <P class="title-area__info-container__info-item__content">{{info.version}}</P>
                </div>
            </VInfo>
            <div class="title-area-divider"></div>
            <div class="title-area__info-container__info-item">
                <p class="title-area__info-container__info-item__title">PERIOD</p>
                <P class="title-area__info-container__info-item__content">{{currentMonth}}</P>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

import VInfo from '@/app/components/VInfo.vue';

import { StatusOnline } from '@/app/store/modules/node';
import { Duration, millisecondsInSecond, minutesInHour, secondsInHour, secondsInMinute } from '@/app/utils/duration';

/**
 * NodeInfo class holds info for NodeInfo entity.
 */
class NodeInfo {
    public id: string;
    public status: string;
    public version: string;
    public allowedVersion: string;
    public wallet: string;
    public isLastVersion: boolean;

    public constructor(id: string, status: string, version: string, allowedVersion: string, wallet: string, isLastVersion: boolean) {
        this.id = id;
        this.status = status;
        this.version = this.toVersionString(version);
        this.allowedVersion = this.toVersionString(allowedVersion);
        this.wallet = wallet;
        this.isLastVersion = isLastVersion;
    }

    private toVersionString(version: string): string {
        return `v${version}`;
    }
}

@Component ({
    components: {
        VInfo,
    },
})
export default class SNOContentTitle extends Vue {
    private timeNow: Date = new Date();

    public mounted(): void {
        window.setInterval(() => {
            this.timeNow = new Date();
        }, 1000);
    }

    public get info(): NodeInfo {
        const nodeInfo = this.$store.state.node.info;

        return new NodeInfo(nodeInfo.id, nodeInfo.status, nodeInfo.version, nodeInfo.allowedVersion, nodeInfo.wallet,
            nodeInfo.isLastVersion);
    }

    public get online(): boolean {
        return this.$store.state.node.info.status === StatusOnline;
    }

    public get lastPinged(): string {
        return this.timePassed(this.$store.state.node.info.lastPinged);
    }

    public get uptime(): string {
        return this.timePassed(this.$store.state.node.info.startedAt);
    }

    public get currentMonth(): string {
        const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
            'July', 'August', 'September', 'October', 'November', 'December'
        ];
        const date = new Date();

        return monthNames[date.getMonth()];
    }

    private timePassed(date: Date): string {
        const difference = Duration.difference(this.timeNow, date);

        if (Math.floor(difference / millisecondsInSecond) > secondsInHour) {
            const hours: string = Math.floor(difference / millisecondsInSecond / secondsInHour) + 'h';
            const minutes: string = Math.floor((difference / millisecondsInSecond % secondsInHour) / minutesInHour) + 'm';

            return `${hours} ${minutes}`;
        }

        return `${Math.floor(difference / millisecondsInSecond / secondsInMinute)}m`;
    }
}
</script>

<style scoped lang="scss">
    .title-area {
        font-family: 'font_regular', sans-serif;
        margin-bottom: 9px;

        &__title {
            font-family: 'font_bold', sans-serif;
            margin: 0 0 21px 0;
            font-size: 32px;
            line-height: 57px;
            color: #535f77;
            user-select: none;
        }

        &__info-container {
            display: flex;
            justify-content: space-between;
            align-items: center;

            &__info-item {
                padding: 15px 0;

                &__title {
                    font-size: 12px;
                    line-height: 20px;
                    color: #9ca5b6;
                    margin: 0 0 5px 0;
                    user-select: none;
                }

                &__content {
                    font-size: 18px;
                    line-height: 20px;
                    font-family: 'font_medium', sans-serif;
                    color: #535f77;
                }
            }
        }
    }

    .title-area-divider {
        width: 1px;
        height: 22px;
        background-color: #dbdfe5;
    }

    .online-status {
        color: #519e62;
    }

    .offline-status {
        color: #ce0000;
    }

    /deep/ .info__message-box {
        background-image: url('../../../static/images/MessageTitle.png');
        bottom: 100%;
        left: 220%;
        padding: 20px 20px 25px 20px;

        &__text {
            align-items: flex-start;

            &__regular-text {
                margin-bottom: 5px;
            }
        }
    }
</style>
